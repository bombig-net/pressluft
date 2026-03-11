<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useDomains } from "~/composables/useDomains";
import { useSites } from "~/composables/useSites";

const {
  domains,
  loading,
  saving,
  error,
  fetchDomains,
  createDomain,
  deleteDomain,
} = useDomains();
const { sites, fetchSites } = useSites();

const pageError = ref("");
const successMessage = ref("");

const form = reactive({
  mode: "attach",
  hostname: "",
  siteId: "",
  assignAsPrimary: true,
});

const baseDomains = computed(() =>
  domains.value.filter((domain) => domain.kind === "base"),
);

const managedDomains = computed(() =>
  domains.value.filter((domain) => domain.kind === "hostname"),
);

const domainStats = computed(() => ({
  total: managedDomains.value.length,
  attached: managedDomains.value.filter((domain) => Boolean(domain.site_id)).length,
  unassigned: managedDomains.value.filter((domain) => !domain.site_id).length,
}));

const siteOptions = computed(() =>
  [...sites.value].sort((a, b) => a.name.localeCompare(b.name)),
);

const canSubmit = computed(() => {
  if (!form.hostname.trim()) {
    return false;
  }
  if (form.mode === "attach") {
    return Boolean(form.siteId);
  }
  return true;
});

const domainKindLabel = (domain: (typeof managedDomains.value)[number]) => {
  if (domain.parent_domain_id || domain.ownership === "platform") {
    return "Temporary URL";
  }
  return "Client domain";
};

const domainRoleLabel = (domain: (typeof managedDomains.value)[number]) =>
  domain.is_primary ? "Primary" : "Additional";

const loadPage = async () => {
  pageError.value = "";
  try {
    await Promise.all([fetchDomains(), fetchSites()]);
    if (!form.siteId && siteOptions.value[0]) {
      form.siteId = siteOptions.value[0].id;
    }
  } catch (e: any) {
    pageError.value = e.message || "Failed to load domains";
  }
};

const resetForm = () => {
  form.mode = "attach";
  form.hostname = "";
  form.siteId = siteOptions.value[0]?.id || "";
  form.assignAsPrimary = true;
};

const handleCreateDomain = async () => {
  pageError.value = "";
  successMessage.value = "";
  try {
    const created = await createDomain({
      hostname: form.hostname,
      site_id: form.mode === "attach" ? form.siteId : undefined,
      is_primary: form.mode === "attach" ? form.assignAsPrimary : false,
    });
    successMessage.value =
      form.mode === "attach" && created.site_name
        ? `Added ${created.hostname} and attached it to ${created.site_name}.`
        : `Added ${created.hostname}.`;
    resetForm();
    await Promise.all([fetchDomains(), fetchSites()]);
  } catch (e: any) {
    pageError.value = e.message || "Failed to add domain";
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
    <section
      class="relative overflow-hidden rounded-[28px] border border-border/60 bg-[linear-gradient(135deg,rgba(255,255,255,0.95),rgba(226,244,240,0.88)_45%,rgba(237,233,254,0.82))] p-7 shadow-[0_32px_120px_-56px_rgba(18,95,84,0.45)] dark:bg-[linear-gradient(135deg,rgba(21,28,31,0.95),rgba(17,57,53,0.88)_50%,rgba(36,33,52,0.88))]"
    >
      <div class="absolute inset-y-0 right-0 hidden w-80 bg-[radial-gradient(circle_at_top,rgba(69,198,214,0.28),transparent_62%)] lg:block" />
      <div class="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-3xl space-y-4">
          <div class="inline-flex items-center gap-2 rounded-full border border-foreground/10 bg-background/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground backdrop-blur">
            Domains
          </div>
          <div>
            <h1 class="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
              Keep every client domain and temporary URL in one place.
            </h1>
            <p class="mt-3 max-w-2xl text-base leading-7 text-muted-foreground">
              Add real customer domains, attach them to sites when needed, and review which address each site currently uses without exposing Pressluft platform setup.
            </p>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-3 sm:min-w-[420px]">
          <div class="rounded-2xl border border-border/60 bg-background/75 px-4 py-4 backdrop-blur">
            <p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">All domains</p>
            <p class="mt-3 text-3xl font-semibold text-foreground">{{ domainStats.total }}</p>
          </div>
          <div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
            <p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-primary/80">Attached</p>
            <p class="mt-3 text-3xl font-semibold text-primary">{{ domainStats.attached }}</p>
          </div>
          <div class="rounded-2xl border border-accent/20 bg-accent/10 px-4 py-4">
            <p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-accent/80">Unassigned</p>
            <p class="mt-3 text-3xl font-semibold text-accent">{{ domainStats.unassigned }}</p>
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
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Add domain</p>
          <h2 class="mt-1 text-xl font-semibold text-foreground">Bring in a domain the agency manages</h2>
        </CardHeader>
        <CardContent class="px-6 py-5">
          <form class="space-y-4" @submit.prevent="handleCreateDomain">
            <div class="flex flex-wrap gap-2">
              <Button type="button" size="sm" :variant="form.mode === 'attach' ? 'default' : 'outline'" @click="form.mode = 'attach'">
                Add and attach
              </Button>
              <Button type="button" size="sm" :variant="form.mode === 'unassigned' ? 'default' : 'outline'" @click="form.mode = 'unassigned'">
                Add only
              </Button>
            </div>
            <div class="space-y-1.5">
              <Label class="text-sm font-medium text-muted-foreground">Domain</Label>
              <Input v-model="form.hostname" placeholder="www.client-example.com" />
            </div>
            <div v-if="form.mode === 'attach'" class="grid gap-4 sm:grid-cols-2">
              <div class="space-y-1.5 sm:col-span-2">
                <Label class="text-sm font-medium text-muted-foreground">Attach to site</Label>
                <select
                  v-model="form.siteId"
                  class="flex h-10 w-full rounded-lg border border-border/60 bg-background/70 px-3 text-sm text-foreground outline-none transition focus:border-accent/40"
                >
                  <option v-for="site in siteOptions" :key="site.id" :value="site.id">
                    {{ site.name }}
                  </option>
                </select>
              </div>
              <label class="flex items-start gap-3 rounded-2xl border border-border/60 bg-muted/20 p-4 text-sm text-muted-foreground sm:col-span-2">
                <input v-model="form.assignAsPrimary" type="checkbox" class="mt-0.5 h-4 w-4 rounded border-border/60" />
                Make this the primary domain for the selected site
              </label>
            </div>
            <div class="rounded-2xl border border-border/60 bg-muted/20 p-4 text-sm leading-6 text-muted-foreground">
              Pressluft-provided temporary URL domains are preinstalled by the platform. This page is for the domains your team brings in and tracks across sites.
            </div>
            <Button type="submit" class="w-full bg-accent text-accent-foreground hover:bg-accent/85" :disabled="saving || !canSubmit">
              {{ saving ? "Adding domain..." : form.mode === "attach" ? "Add and attach domain" : "Add domain" }}
            </Button>
          </form>
        </CardContent>
      </Card>

      <div class="space-y-6">
        <Card class="rounded-[24px] border border-border/60 bg-card/70 py-0 shadow-none">
          <CardHeader class="border-b border-border/50 px-6 py-5">
            <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Inventory</p>
            <h2 class="mt-1 text-xl font-semibold text-foreground">All tracked domains</h2>
          </CardHeader>
          <CardContent class="px-6 py-5">
            <div v-if="loading" class="py-10 text-sm text-muted-foreground">Loading domains...</div>
            <div v-else-if="managedDomains.length === 0" class="rounded-2xl border border-dashed border-border/60 bg-muted/20 px-4 py-8 text-center text-sm text-muted-foreground">
               No domains tracked yet.
            </div>
            <div v-else class="space-y-3">
              <div v-for="domain in managedDomains" :key="domain.id" class="rounded-2xl border border-border/60 bg-background/70 p-4">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="text-sm font-semibold text-foreground">{{ domain.hostname }}</p>
                      <Badge variant="outline" class="border-primary/30 bg-primary/10 text-primary">{{ domainRoleLabel(domain) }}</Badge>
                      <Badge variant="outline" class="border-border/60 bg-muted/40 text-muted-foreground">{{ domainKindLabel(domain) }}</Badge>
                    </div>
                    <p class="mt-1 text-xs text-muted-foreground">
                      {{ domain.site_name ? `Attached to ${domain.site_name}` : "Unassigned" }}
                      <span v-if="domain.parent_hostname"> · via {{ domain.parent_hostname }}</span>
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

        <Card v-if="baseDomains.length > 0" class="rounded-[24px] border border-border/60 bg-card/70 py-0 shadow-none">
          <CardHeader class="border-b border-border/50 px-6 py-5">
            <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Pressluft URLs</p>
            <h2 class="mt-1 text-xl font-semibold text-foreground">Platform-provided temporary domains</h2>
          </CardHeader>
          <CardContent class="space-y-3 px-6 py-5 text-sm text-muted-foreground">
            <p>
              These are preinstalled by Pressluft and used when a site gets a temporary URL. They are shown here for context, not as something operators need to manage day to day.
            </p>
            <div class="flex flex-wrap gap-2">
              <Badge v-for="domain in baseDomains" :key="domain.id" variant="outline" class="border-border/60 bg-muted/40 text-muted-foreground">
                {{ domain.hostname }}
              </Badge>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
