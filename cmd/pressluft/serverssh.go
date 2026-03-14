package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"pressluft/internal/security"

	_ "modernc.org/sqlite"
)

type sshAccessTarget struct {
	ID   string
	Name string
	IPv4 string
	IPv6 string
	Key  string
}

func runServerSSH(args []string) error {
	target, execSSH, printKey, err := parseServerSSHArgs(args)
	if err != nil {
		return err
	}

	runtime, err := resolveRuntime()
	if err != nil {
		return fmt.Errorf("resolve runtime: %w", err)
	}

	db, err := openExistingDB(runtime.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	access, err := lookupSSHAccessTarget(db, target)
	if err != nil {
		return err
	}
	decryptedKey, err := security.Decrypt(access.Key)
	if err != nil {
		return fmt.Errorf("decrypt server ssh key: %w", err)
	}

	if printKey {
		_, err := os.Stdout.Write(decryptedKey)
		return err
	}

	host := strings.TrimSpace(access.IPv4)
	if host == "" {
		host = strings.TrimSpace(access.IPv6)
	}
	if host == "" {
		return fmt.Errorf("server %q has no recorded IP address", access.ID)
	}

	keyFile, err := os.CreateTemp("", "pressluft-server-ssh-*.key")
	if err != nil {
		return fmt.Errorf("create temp key file: %w", err)
	}
	defer keyFile.Close()
	if err := keyFile.Chmod(0o600); err != nil {
		return fmt.Errorf("chmod temp key file: %w", err)
	}
	if _, err := keyFile.Write(decryptedKey); err != nil {
		return fmt.Errorf("write temp key file: %w", err)
	}

	sshArgs := []string{"-i", keyFile.Name(), "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "root@" + host}
	if execSSH {
		cmd := exec.Command("ssh", sshArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("run ssh: %w", err)
		}
		fmt.Fprintf(os.Stderr, "\nSSH key left at %s\n", keyFile.Name())
		return nil
	}

	fmt.Printf("server: %s (%s)\n", access.Name, access.ID)
	fmt.Printf("host: %s\n", host)
	fmt.Printf("key_file: %s\n", keyFile.Name())
	fmt.Printf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@%s\n", keyFile.Name(), host)
	return nil
}

func parseServerSSHArgs(args []string) (target string, execSSH bool, printKey bool, err error) {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("pressluft server-ssh TARGET — print or execute SSH access for a managed server")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  --exec       Open an interactive SSH session")
			fmt.Println("  --print-key  Print the decrypted SSH private key to stdout")
			os.Exit(0)
		}
	}
	for _, arg := range args {
		switch arg {
		case "--exec":
			execSSH = true
		case "--print-key":
			printKey = true
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, false, fmt.Errorf("unknown server-ssh argument %q", arg)
			}
			if target != "" {
				return "", false, false, fmt.Errorf("server-ssh accepts exactly one TARGET")
			}
			target = strings.TrimSpace(arg)
		}
	}
	if target == "" {
		return "", false, false, fmt.Errorf("server-ssh requires TARGET")
	}
	if execSSH && printKey {
		return "", false, false, fmt.Errorf("--exec and --print-key cannot be combined")
	}
	return target, execSSH, printKey, nil
}

func lookupSSHAccessTarget(db *sql.DB, target string) (*sshAccessTarget, error) {
	rows, err := db.Query(`
		SELECT s.id, s.name, s.ipv4, s.ipv6, k.private_key_encrypted
		FROM servers s
		JOIN server_keys k ON k.server_id = s.id
		WHERE s.id = ? OR s.name = ?
		ORDER BY s.created_at DESC
	`, target, target)
	if err != nil {
		return nil, fmt.Errorf("query server ssh access: %w", err)
	}
	defer rows.Close()

	var matches []sshAccessTarget
	for rows.Next() {
		var item sshAccessTarget
		var ipv4 sql.NullString
		var ipv6 sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &ipv4, &ipv6, &item.Key); err != nil {
			return nil, fmt.Errorf("scan server ssh access: %w", err)
		}
		if ipv4.Valid {
			item.IPv4 = ipv4.String
		}
		if ipv6.Valid {
			item.IPv6 = ipv6.String
		}
		matches = append(matches, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate server ssh access: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no server with stored SSH key matched %q", target)
	}
	if len(matches) > 1 {
		ids := make([]string, 0, len(matches))
		for _, match := range matches {
			ids = append(ids, fmt.Sprintf("%s (%s)", match.Name, match.ID))
		}
		return nil, fmt.Errorf("multiple servers matched %q: %s", target, strings.Join(ids, ", "))
	}
	return &matches[0], nil
}

func openExistingDB(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("stat db: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
