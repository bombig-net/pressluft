<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useDomains } from "~/composables/useDomains";

const {
  domains,
  loading,
  saving,
  error,
  fetchDomains,
  createDomain,
  deleteDomain,
} = useDomains();

const pageError = ref("");
const successMessage = ref("");

const form = reactive({
  hostname: "",
});

const baseDomains = computed(() =>
  domains.value.filter((domain) => domain.kind === "base"),
);

const assignedHostnames = computed(() =>
  domains.value.filter((domain) => domain.kind === "hostname"),
);

const loadPage = async () => {
  pageError.value = "";
  try {
    await fetchDomains();
  } catch (e: any) {
    pageError.value = e.message || "Failed to load domains";
  }
};

const handleCreateBaseDomain = async () => {
  pageError.value = "";
  successMessage.value = "";
  try {
    const created = await createDomain({ hostname: form.hostname });
    successMessage.value = `Added ${created.hostname} to the sandbox domain inventory.`;
    form.hostname = "";
    await fetchDomains();
  } catch (e: any) {
    pageError.value = e.message || "Failed to create base domain";
  }
};

const handleDelete = async (domainId: string, hostname: string) => {
  if (!window.confirm(`Delete ${hostname}?`)) {
    return;
  }
  pageError.value = "";
  successMessage.value = "";
  try {
    await deleteDomain(domainId);
    successMessage.value = `Deleted ${hostname}.`;
    await fetchDomains();
  } catch (e: any) {
    pageError.value = e.message || "Failed to delete domain";
  }
};

onMounted(loadPage);
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-[28px] border border-border/60 bg-[linear-gradient(135deg,rgba(255,248,239,0.96),rgba(237,248,244,0.9)_45%,rgba(236,243,255,0.9))] p-7 shadow-[0_32px_120px_-56px_rgba(32,88,74,0.35)]">
      <div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-3xl space-y-4">
          <div class="inline-flex items-center gap-2 rounded-full border border-foreground/10 bg-background/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground backdrop-blur">
            Domains Foundation
          </div>
          <div>
            <h1 class="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              Hostnames now live as their own platform inventory.
            </h1>
            <p class="mt-3 max-w-2xl text-base leading-7 text-muted-foreground">
              Manage Pressluft-owned base domains, review assigned hostnames, and keep the control plane ready for sandbox domains now and custom domains later.
            </p>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3 sm:min-w-[320px]">
          <div class="rounded-2xl border border-border/60 bg-background/75 px-4 py-4 backdrop-blur">
            <p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">Base domains</p>
            <p class="mt-3 text-3xl font-semibold text-foreground">{{ baseDomains.length }}</p>
          </div>
          <div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
            <p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-primary/80">Assigned hostnames</p>
            <p class="mt-3 text-3xl font-semibold text-primary">{{ assignedHostnames.length }}</p>
          </div>
        </div>
      </div>
    </section>

    <Alert v-if="pageError || error" class="border-destructive/30 bg-destructive/10 text-destructive">
      <AlertDescription>{{ pageError || error }}</AlertDescription>
    </Alert>
    <Alert v-if="successMessage" class="border-primary/30 bg-primary/10 text-primary">
      <AlertDescription>{{ successMessage }}</AlertDescription>
    </Alert>

    <div class="grid gap-6 xl:grid-cols-[0.95fr_1.25fr]">
      <Card class="rounded-[24px] border border-border/60 bg-card/70 py-0 shadow-none">
        <CardHeader class="border-b border-border/50 px-6 py-5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">New base domain</p>
          <h2 class="mt-1 text-xl font-semibold text-foreground">Add sandbox inventory</h2>
        </CardHeader>
        <CardContent class="px-6 py-5">
          <form class="space-y-4" @submit.prevent="handleCreateBaseDomain">
            <div class="space-y-1.5">
              <Label class="text-sm font-medium text-muted-foreground">Base domain</Label>
              <Input v-model="form.hostname" placeholder="sandbox.pressluft.test" />
            </div>
            <div class="rounded-2xl border border-border/60 bg-muted/20 p-4 text-sm leading-6 text-muted-foreground">
              Base domains stay platform-managed and become the parent pool for generated sandbox hostnames on site records.
            </div>
            <Button type="submit" class="w-full bg-accent text-accent-foreground hover:bg-accent/85" :disabled="saving || !form.hostname.trim()">
              {{ saving ? "Adding domain..." : "Add base domain" }}
            </Button>
          </form>
        </CardContent>
      </Card>

      <div class="space-y-6">
        <Card class="rounded-[24px] border border-border/60 bg-card/70 py-0 shadow-none">
          <CardHeader class="border-b border-border/50 px-6 py-5">
            <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Platform inventory</p>
            <h2 class="mt-1 text-xl font-semibold text-foreground">Base domains</h2>
          </CardHeader>
          <CardContent class="px-6 py-5">
            <div v-if="loading" class="py-10 text-sm text-muted-foreground">Loading domains...</div>
            <div v-else-if="baseDomains.length === 0" class="rounded-2xl border border-dashed border-border/60 bg-muted/20 px-4 py-8 text-center text-sm text-muted-foreground">
              No base domains configured yet.
            </div>
            <div v-else class="space-y-3">
              <div v-for="domain in baseDomains" :key="domain.id" class="rounded-2xl border border-border/60 bg-background/70 p-4">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="text-sm font-semibold text-foreground">{{ domain.hostname }}</p>
                      <Badge variant="outline" class="border-primary/30 bg-primary/10 text-primary">{{ domain.source }}</Badge>
                    </div>
                    <p class="mt-1 text-xs text-muted-foreground">Pressluft-owned parent domain for sandbox hostname generation.</p>
                  </div>
                  <Button type="button" variant="ghost" size="sm" class="text-destructive hover:bg-destructive/10 hover:text-destructive" @click="handleDelete(domain.id, domain.hostname)">
                    Delete
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card class="rounded-[24px] border border-border/60 bg-card/70 py-0 shadow-none">
          <CardHeader class="border-b border-border/50 px-6 py-5">
            <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Assignments</p>
            <h2 class="mt-1 text-xl font-semibold text-foreground">Hostname inventory</h2>
          </CardHeader>
          <CardContent class="px-6 py-5">
            <div v-if="assignedHostnames.length === 0" class="rounded-2xl border border-dashed border-border/60 bg-muted/20 px-4 py-8 text-center text-sm text-muted-foreground">
              No hostnames assigned yet.
            </div>
            <div v-else class="space-y-3">
              <div v-for="domain in assignedHostnames" :key="domain.id" class="rounded-2xl border border-border/60 bg-background/70 p-4">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="text-sm font-semibold text-foreground">{{ domain.hostname }}</p>
                      <Badge v-if="domain.is_primary" variant="outline" class="border-primary/30 bg-primary/10 text-primary">Primary</Badge>
                      <Badge variant="outline" class="border-border/60 bg-muted/40 text-muted-foreground">{{ domain.ownership }}</Badge>
                    </div>
                    <p class="mt-1 text-xs text-muted-foreground">
                      {{ domain.site_name ? `Assigned to ${domain.site_name}` : "Not assigned to a site" }}
                      <span v-if="domain.parent_hostname"> · {{ domain.parent_hostname }}</span>
                    </p>
                  </div>
                  <div class="flex items-center gap-3 text-xs text-muted-foreground">
                    <NuxtLink v-if="domain.site_id" :to="`/sites/${domain.site_id}`" class="font-medium text-accent transition hover:text-accent/80">
                      Open site
                    </NuxtLink>
                    <Button type="button" variant="ghost" size="sm" class="text-destructive hover:bg-destructive/10 hover:text-destructive" @click="handleDelete(domain.id, domain.hostname)">
                      Delete
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
