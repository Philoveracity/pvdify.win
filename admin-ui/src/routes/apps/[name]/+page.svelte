<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';

  interface App {
    name: string;
    status: string;
    environment: string;
    image: string;
    bind_port: number;
    domains: string[];
    created_at: string;
    updated_at: string;
  }

  interface Release {
    version: number;
    image: string;
    status: string;
    created_at: string;
  }

  interface Process {
    type: string;
    count: number;
    status: string;
    command?: string;
  }

  let app: App | null = null;
  let releases: Release[] = [];
  let processes: Process[] = [];
  let config: Record<string, string> = {};
  let loading = true;
  let activeTab = 'overview';
  let showConfigValues = false;
  let mobileActionsOpen = false;

  const API_URL = import.meta.env.VITE_API_URL || '';

  $: if ($page.params.name) {
    loadApp($page.params.name);
  }

  async function loadApp(name: string) {
    loading = true;
    try {
      const [appRes, releasesRes, psRes, configRes] = await Promise.all([
        fetch(`${API_URL}/api/v1/apps/${name}`),
        fetch(`${API_URL}/api/v1/apps/${name}/releases`),
        fetch(`${API_URL}/api/v1/apps/${name}/ps`),
        fetch(`${API_URL}/api/v1/apps/${name}/config`)
      ]);

      // API returns {app: {...}, domains: [...], processes: [...]}
      const appData = await appRes.json();
      app = appData.app || appData;
      if (appData.domains) {
        app.domains = appData.domains;
      }

      // Releases is an array
      releases = await releasesRes.json().catch(() => []);

      // PS returns {definitions: [...], instances: [...]}
      const psData = await psRes.json().catch(() => ({ definitions: [] }));
      processes = (psData.definitions || []).map((def: any) => ({
        type: def.name,
        count: def.count,
        status: 'running',
        command: def.command
      }));

      // Config returns {vars: {...}, version: ...}
      const configData = await configRes.json().catch(() => ({ vars: {} }));
      config = configData.vars || configData || {};
    } catch (e) {
      console.error('Error loading app:', e);
    } finally {
      loading = false;
    }
  }

  function getStatusConfig(status: string) {
    switch (status) {
      case 'running': return { color: 'status-success', bg: 'bg-green-50', text: 'text-green-700', label: 'Running' };
      case 'stopped': return { color: 'status-neutral', bg: 'bg-gray-50', text: 'text-gray-600', label: 'Stopped' };
      case 'deploying': return { color: 'status-warning', bg: 'bg-yellow-50', text: 'text-yellow-700', label: 'Deploying' };
      case 'failed': return { color: 'status-error', bg: 'bg-red-50', text: 'text-red-700', label: 'Failed' };
      case 'pending': return { color: 'status-warning', bg: 'bg-yellow-50', text: 'text-yellow-700', label: 'Pending' };
      default: return { color: 'status-neutral', bg: 'bg-gray-50', text: 'text-gray-600', label: status || 'Unknown' };
    }
  }

  function formatDate(date: string) {
    return new Date(date).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function formatRelativeDate(date: string) {
    const d = new Date(date);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return formatDate(date);
  }

  const tabs = [
    { id: 'overview', label: 'Overview', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z' },
    { id: 'deploy', label: 'Deploy', icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12' },
    { id: 'config', label: 'Config', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' },
    { id: 'settings', label: 'Settings', icon: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4' }
  ];
</script>

<svelte:head>
  <title>{app?.name || 'Loading...'} | Pvdify</title>
</svelte:head>

{#if loading}
  <!-- Skeleton Loading State -->
  <div class="bg-white border-b border-gray-200">
    <div class="page-container py-6">
      <div class="flex items-center gap-4">
        <div class="skeleton w-4 h-4 rounded-full"></div>
        <div>
          <div class="skeleton h-7 w-40 mb-2"></div>
          <div class="skeleton h-5 w-56"></div>
        </div>
      </div>
    </div>
    <div class="page-container">
      <div class="flex gap-4 pb-3">
        {#each [1, 2, 3, 4] as _}
          <div class="skeleton h-10 w-24"></div>
        {/each}
      </div>
    </div>
  </div>
  <div class="page-container py-6">
    <div class="grid gap-6">
      <div class="card p-6">
        <div class="skeleton h-5 w-32 mb-4"></div>
        <div class="skeleton h-20 w-full"></div>
      </div>
    </div>
  </div>

{:else if app}
  <!-- App Header -->
  <div class="bg-white border-b border-gray-200 sticky top-16 z-30">
    <div class="page-container">
      <!-- Main Header Row -->
      <div class="py-4 sm:py-6">
        <div class="flex items-start sm:items-center justify-between gap-4">
          <!-- Back + App Info -->
          <div class="flex items-start sm:items-center gap-3 sm:gap-4 min-w-0">
            <a href="/" class="flex-shrink-0 p-2 -ml-2 rounded-lg hover:bg-gray-100 transition-colors sm:hidden">
              <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
              </svg>
            </a>

            <div class="hidden sm:block flex-shrink-0">
              <span class="block w-4 h-4 rounded-full {getStatusConfig(app.status).color} {app.status === 'deploying' ? 'animate-pulse' : ''}"></span>
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 flex-wrap">
                <h1 class="text-lg sm:text-xl font-semibold text-gray-900 truncate">{app.name}</h1>
                <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium {getStatusConfig(app.status).bg} {getStatusConfig(app.status).text}">
                  <span class="w-1.5 h-1.5 rounded-full {getStatusConfig(app.status).color} {app.status === 'deploying' ? 'animate-pulse' : ''} sm:hidden"></span>
                  {getStatusConfig(app.status).label}
                </span>
              </div>
              <a
                href="https://{app.domains?.[0] || `${app.name}.pvdify.win`}"
                target="_blank"
                rel="noopener noreferrer"
                class="text-sm text-gray-500 hover:text-pvd-600 transition-colors truncate block mt-0.5"
              >
                {app.domains?.[0] || `${app.name}.pvdify.win`}
              </a>
            </div>
          </div>

          <!-- Desktop Actions -->
          <div class="hidden sm:flex items-center gap-2 flex-shrink-0">
            <a
              href="https://{app.domains?.[0] || `${app.name}.pvdify.win`}"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary btn-sm gap-1.5"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
              </svg>
              Open
            </a>
          </div>

          <!-- Mobile Actions Toggle -->
          <button
            type="button"
            on:click|stopPropagation={() => mobileActionsOpen = !mobileActionsOpen}
            class="sm:hidden p-2 -mr-2 rounded-lg hover:bg-gray-100 transition-colors touch-target"
          >
            <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"/>
            </svg>
          </button>
        </div>

        <!-- Mobile Actions Dropdown -->
        {#if mobileActionsOpen}
          <div class="sm:hidden mt-3 pt-3 border-t border-gray-100 flex gap-2">
            <a
              href="https://{app.domains?.[0] || `${app.name}.pvdify.win`}"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary btn-sm flex-1 justify-center gap-1.5"
              on:click={() => mobileActionsOpen = false}
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
              </svg>
              Open App
            </a>
          </div>
        {/if}
      </div>

      <!-- Tabs - Horizontally Scrollable on Mobile -->
      <nav class="tabs -mb-px">
        {#each tabs as tab}
          <button
            type="button"
            on:click|stopPropagation={() => activeTab = tab.id}
            class="tab touch-target {activeTab === tab.id ? 'tab-active' : ''}"
          >
            <svg class="w-4 h-4 sm:hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={tab.icon}/>
            </svg>
            <span>{tab.label}</span>
          </button>
        {/each}
      </nav>
    </div>
  </div>

  <!-- Tab Content -->
  <div class="page-container py-6">
    <!-- Overview Tab -->
    {#if activeTab === 'overview'}
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-6">
        <!-- Dynos Card -->
        <div class="lg:col-span-2">
          <div class="card">
            <div class="card-header">
              <h2 class="font-medium text-gray-900">Dynos</h2>
              <span class="text-sm text-gray-500">{processes.length} process{processes.length !== 1 ? 'es' : ''}</span>
            </div>
            {#if processes.length}
              <div class="divide-y divide-gray-100">
                {#each processes as proc}
                  <div class="px-4 sm:px-6 py-4 flex items-center justify-between gap-4">
                    <div class="flex items-center gap-3 min-w-0">
                      <div class="w-10 h-10 rounded-lg bg-pvd-50 flex items-center justify-center flex-shrink-0">
                        <svg class="w-5 h-5 text-pvd-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7"/>
                        </svg>
                      </div>
                      <div class="min-w-0">
                        <div class="font-medium text-gray-900">{proc.type}</div>
                        <div class="text-sm text-gray-500 truncate">{proc.command || 'Default command'}</div>
                      </div>
                    </div>
                    <div class="flex items-center gap-3 flex-shrink-0">
                      <span class="badge badge-success">{proc.count}x</span>
                      <button class="btn btn-ghost btn-sm hidden sm:inline-flex">Scale</button>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="px-4 sm:px-6 py-8 text-center">
                <div class="w-12 h-12 mx-auto mb-3 rounded-full bg-gray-100 flex items-center justify-center">
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
                  </svg>
                </div>
                <p class="text-sm text-gray-500 mb-4">No processes running</p>
                <button on:click={() => activeTab = 'deploy'} class="btn btn-primary btn-sm">
                  Deploy an Image
                </button>
              </div>
            {/if}
          </div>
        </div>

        <!-- Latest Release Card -->
        <div class="card">
          <div class="card-header">
            <h2 class="font-medium text-gray-900">Latest Release</h2>
          </div>
          {#if releases.length}
            {@const latest = releases[0]}
            {@const releaseStatus = getStatusConfig(latest.status)}
            <div class="px-4 sm:px-6 py-4">
              <div class="flex items-center justify-between mb-3">
                <span class="badge bg-gray-100 text-gray-700">v{latest.version}</span>
                <span class="badge {releaseStatus.bg} {releaseStatus.text}">{releaseStatus.label}</span>
              </div>
              <div class="text-sm text-gray-600 font-mono truncate mb-2" title={latest.image}>
                {latest.image}
              </div>
              <div class="text-xs text-gray-400">
                {formatRelativeDate(latest.created_at)}
              </div>
            </div>
            <div class="px-4 sm:px-6 py-3 border-t border-gray-100 bg-gray-50/50">
              <button on:click={() => activeTab = 'deploy'} class="text-sm text-pvd-600 hover:text-pvd-700 font-medium">
                View all releases →
              </button>
            </div>
          {:else}
            <div class="px-4 sm:px-6 py-8 text-center">
              <p class="text-sm text-gray-500">No releases yet</p>
            </div>
          {/if}
        </div>

        <!-- Quick Stats -->
        <div class="lg:col-span-3 grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
          <div class="card card-stats">
            <div class="text-2xl font-semibold text-gray-900">{releases.length}</div>
            <div class="text-xs text-gray-500 mt-1">Total Releases</div>
          </div>
          <div class="card card-stats">
            <div class="text-2xl font-semibold text-gray-900">{Object.keys(config).length}</div>
            <div class="text-xs text-gray-500 mt-1">Config Vars</div>
          </div>
          <div class="card card-stats">
            <div class="text-2xl font-semibold text-gray-900">{app.domains?.length || 1}</div>
            <div class="text-xs text-gray-500 mt-1">Domain{(app.domains?.length || 1) !== 1 ? 's' : ''}</div>
          </div>
          <div class="card card-stats">
            <div class="text-2xl font-semibold text-gray-900">{processes.reduce((sum, p) => sum + p.count, 0)}</div>
            <div class="text-xs text-gray-500 mt-1">Dyno{processes.reduce((sum, p) => sum + p.count, 0) !== 1 ? 's' : ''}</div>
          </div>
        </div>
      </div>

    <!-- Deploy Tab -->
    {:else if activeTab === 'deploy'}
      <div class="space-y-6">
        <!-- Deploy CLI Card -->
        <div class="card">
          <div class="card-header">
            <h2 class="font-medium text-gray-900">Deploy via CLI</h2>
          </div>
          <div class="px-4 sm:px-6 py-4 sm:py-6">
            <p class="text-sm text-gray-500 mb-4">Deploy a container image to your app using the Pvdify CLI:</p>
            <div class="bg-gray-900 rounded-lg p-4 overflow-x-auto">
              <code class="text-sm text-gray-100 whitespace-nowrap">pvdify deploy {app.name} --image your-image:tag</code>
            </div>
            <p class="text-xs text-gray-400 mt-3">
              Or use <code class="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">gh pvdify deploy</code> for GitHub integration
            </p>
          </div>
        </div>

        <!-- Release History -->
        <div class="card">
          <div class="card-header">
            <h2 class="font-medium text-gray-900">Release History</h2>
            <span class="text-sm text-gray-500">{releases.length} release{releases.length !== 1 ? 's' : ''}</span>
          </div>
          {#if releases.length}
            <div class="divide-y divide-gray-100">
              {#each releases as release, i}
                {@const releaseStatus = getStatusConfig(release.status)}
                <div class="px-4 sm:px-6 py-4 flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4">
                  <div class="flex items-center gap-3 flex-1 min-w-0">
                    <span class="badge bg-gray-100 text-gray-700 flex-shrink-0">v{release.version}</span>
                    <span class="text-sm text-gray-900 font-mono truncate" title={release.image}>{release.image}</span>
                  </div>
                  <div class="flex items-center justify-between sm:justify-end gap-3 flex-shrink-0">
                    <span class="badge {releaseStatus.bg} {releaseStatus.text}">{releaseStatus.label}</span>
                    <span class="text-sm text-gray-400">{formatRelativeDate(release.created_at)}</span>
                    {#if release.version > 1 && i > 0}
                      <button class="btn btn-ghost btn-sm hidden sm:inline-flex">Rollback</button>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="px-4 sm:px-6 py-8 text-center text-gray-500">
              No releases yet. Deploy your first image to get started.
            </div>
          {/if}
        </div>
      </div>

    <!-- Config Tab -->
    {:else if activeTab === 'config'}
      <div class="card">
        <div class="card-header">
          <h2 class="font-medium text-gray-900">Config Vars</h2>
          <button
            on:click={() => showConfigValues = !showConfigValues}
            class="btn btn-ghost btn-sm gap-1.5"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {#if showConfigValues}
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/>
              {:else}
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
              {/if}
            </svg>
            {showConfigValues ? 'Hide' : 'Reveal'}
          </button>
        </div>
        {#if Object.keys(config).length}
          <div class="divide-y divide-gray-100">
            {#each Object.entries(config) as [key, value]}
              <div class="px-4 sm:px-6 py-3 flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-4">
                <div class="font-mono text-sm text-gray-700 font-medium sm:w-48 flex-shrink-0">{key}</div>
                <div class="font-mono text-sm text-gray-500 truncate flex-1">
                  {#if showConfigValues}
                    {value}
                  {:else}
                    ••••••••
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="px-4 sm:px-6 py-8 text-center">
            <div class="w-12 h-12 mx-auto mb-3 rounded-full bg-gray-100 flex items-center justify-center">
              <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/>
              </svg>
            </div>
            <p class="text-sm text-gray-500 mb-4">No config vars set</p>
            <p class="text-xs text-gray-400">
              Use <code class="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">pvdify config:set KEY=value</code> to add
            </p>
          </div>
        {/if}
      </div>

    <!-- Settings Tab -->
    {:else if activeTab === 'settings'}
      <div class="space-y-6">
        <!-- App Information -->
        <div class="card">
          <div class="card-header">
            <h2 class="font-medium text-gray-900">App Information</h2>
          </div>
          <div class="px-4 sm:px-6 py-4 sm:py-6">
            <dl class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
              <div>
                <dt class="text-sm text-gray-500 mb-1">Name</dt>
                <dd class="font-medium text-gray-900">{app.name}</dd>
              </div>
              <div>
                <dt class="text-sm text-gray-500 mb-1">Environment</dt>
                <dd class="font-medium text-gray-900">{app.environment || 'production'}</dd>
              </div>
              <div>
                <dt class="text-sm text-gray-500 mb-1">Created</dt>
                <dd class="font-medium text-gray-900">{formatDate(app.created_at)}</dd>
              </div>
              <div>
                <dt class="text-sm text-gray-500 mb-1">Last Updated</dt>
                <dd class="font-medium text-gray-900">{formatDate(app.updated_at)}</dd>
              </div>
              {#if app.image}
                <div class="sm:col-span-2">
                  <dt class="text-sm text-gray-500 mb-1">Current Image</dt>
                  <dd class="font-mono text-sm text-gray-900 truncate">{app.image}</dd>
                </div>
              {/if}
            </dl>
          </div>
        </div>

        <!-- Domains -->
        <div class="card">
          <div class="card-header">
            <h2 class="font-medium text-gray-900">Domains</h2>
          </div>
          <div class="divide-y divide-gray-100">
            <!-- Default Domain -->
            <div class="px-4 sm:px-6 py-4 flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
                  <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
                  </svg>
                </div>
                <div class="min-w-0">
                  <a
                    href="https://{app.name}.pvdify.win"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-sm text-pvd-600 hover:text-pvd-700 truncate block"
                  >
                    {app.name}.pvdify.win
                  </a>
                  <span class="text-xs text-gray-400">Default domain</span>
                </div>
              </div>
              <span class="badge badge-success flex-shrink-0">Active</span>
            </div>

            <!-- Custom Domains -->
            {#if app.domains?.length}
              {#each app.domains as domain}
                <div class="px-4 sm:px-6 py-4 flex items-center justify-between gap-4">
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="w-8 h-8 rounded-lg bg-pvd-50 flex items-center justify-center flex-shrink-0">
                      <svg class="w-4 h-4 text-pvd-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
                      </svg>
                    </div>
                    <a
                      href="https://{domain}"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="text-sm text-pvd-600 hover:text-pvd-700 truncate"
                    >
                      {domain}
                    </a>
                  </div>
                  <span class="badge badge-success flex-shrink-0">Active</span>
                </div>
              {/each}
            {/if}
          </div>
          <div class="px-4 sm:px-6 py-3 border-t border-gray-100 bg-gray-50/50">
            <p class="text-xs text-gray-500">
              Add custom domains via CLI: <code class="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">pvdify domains:add {app.name} example.com</code>
            </p>
          </div>
        </div>

        <!-- Danger Zone -->
        <div class="card border-red-200 overflow-hidden">
          <div class="px-4 sm:px-6 py-4 bg-red-50 border-b border-red-200">
            <h2 class="font-medium text-red-900">Danger Zone</h2>
          </div>
          <div class="px-4 sm:px-6 py-4 sm:py-6">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h3 class="font-medium text-gray-900">Delete this app</h3>
                <p class="text-sm text-gray-500 mt-1">Permanently delete {app.name} and all its data. This cannot be undone.</p>
              </div>
              <button class="btn btn-danger flex-shrink-0 w-full sm:w-auto justify-center">
                Delete App
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
{/if}
