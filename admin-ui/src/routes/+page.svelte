<script lang="ts">
  import { onMount } from 'svelte';

  interface App {
    name: string;
    status: string;
    image: string;
    domains: string[];
    created_at: string;
    updated_at: string;
  }

  let apps: App[] = [];
  let loading = true;
  let error: string | null = null;

  const API_URL = import.meta.env.VITE_API_URL || '';

  onMount(async () => {
    try {
      const res = await fetch(`${API_URL}/api/v1/apps`);
      if (!res.ok) throw new Error('Failed to fetch apps');
      apps = await res.json();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error';
    } finally {
      loading = false;
    }
  });

  function getStatusConfig(status: string) {
    switch (status) {
      case 'running': return { color: 'status-success', label: 'Running', icon: '●' };
      case 'stopped': return { color: 'status-neutral', label: 'Stopped', icon: '○' };
      case 'deploying': return { color: 'status-warning', label: 'Deploying', icon: '◐' };
      case 'failed': return { color: 'status-error', label: 'Failed', icon: '✕' };
      default: return { color: 'status-neutral', label: status || 'Unknown', icon: '○' };
    }
  }

  function formatDate(date: string) {
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

    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>Dashboard | Pvdify</title>
</svelte:head>

<div class="page-container py-6 sm:py-8">
  <!-- Page Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
    <div>
      <h1 class="text-xl sm:text-2xl font-semibold text-gray-900">Your Apps</h1>
      <p class="mt-1 text-sm text-gray-500">Deploy and manage containerized applications</p>
    </div>
    <a href="/apps/new" class="btn btn-primary gap-2 w-full sm:w-auto justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
      </svg>
      <span>Create App</span>
    </a>
  </div>

  <!-- Stats Bar (visible when apps exist) -->
  {#if !loading && !error && apps.length > 0}
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4 mb-6">
      <div class="card card-stats">
        <div class="text-2xl font-semibold text-gray-900">{apps.length}</div>
        <div class="text-xs text-gray-500 mt-1">Total Apps</div>
      </div>
      <div class="card card-stats">
        <div class="text-2xl font-semibold text-green-600">{apps.filter(a => a.status === 'running').length}</div>
        <div class="text-xs text-gray-500 mt-1">Running</div>
      </div>
      <div class="card card-stats">
        <div class="text-2xl font-semibold text-yellow-600">{apps.filter(a => a.status === 'deploying').length}</div>
        <div class="text-xs text-gray-500 mt-1">Deploying</div>
      </div>
      <div class="card card-stats">
        <div class="text-2xl font-semibold text-gray-400">{apps.filter(a => a.status === 'stopped').length}</div>
        <div class="text-xs text-gray-500 mt-1">Stopped</div>
      </div>
    </div>
  {/if}

  <!-- Loading State with Skeleton -->
  {#if loading}
    <div class="card overflow-hidden">
      {#each [1, 2, 3] as _}
        <div class="px-4 sm:px-6 py-4 border-b border-gray-100 last:border-0">
          <div class="flex items-center gap-4">
            <div class="skeleton w-3 h-3 rounded-full"></div>
            <div class="flex-1 min-w-0">
              <div class="skeleton h-5 w-32 mb-2"></div>
              <div class="skeleton h-4 w-48"></div>
            </div>
            <div class="hidden sm:block">
              <div class="skeleton h-4 w-20"></div>
            </div>
          </div>
        </div>
      {/each}
    </div>

  <!-- Error State -->
  {:else if error}
    <div class="card p-8 sm:p-12 text-center">
      <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center">
        <svg class="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Unable to load apps</h3>
      <p class="text-sm text-gray-500 mb-6 max-w-sm mx-auto">{error}</p>
      <button on:click={() => location.reload()} class="btn btn-secondary">
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
        </svg>
        Try Again
      </button>
    </div>

  <!-- Empty State -->
  {:else if apps.length === 0}
    <div class="card p-8 sm:p-12 text-center">
      <div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-pvd-50 to-pvd-100 flex items-center justify-center">
        <svg class="w-10 h-10 text-pvd-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Create your first app</h3>
      <p class="text-sm text-gray-500 mb-6 max-w-sm mx-auto">
        Deploy containerized applications with zero configuration. Just push your image and go.
      </p>
      <a href="/apps/new" class="btn btn-primary inline-flex">
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        Create App
      </a>

      <!-- Quick start hint -->
      <div class="mt-8 pt-6 border-t border-gray-100">
        <p class="text-xs text-gray-400 mb-3">Or deploy via CLI</p>
        <code class="text-xs bg-gray-900 text-gray-100 px-3 py-2 rounded-lg inline-block">
          pvdify apps:create my-app
        </code>
      </div>
    </div>

  <!-- App List -->
  {:else}
    <div class="card overflow-hidden divide-y divide-gray-100">
      {#each apps as app}
        {@const statusConfig = getStatusConfig(app.status)}
        <a
          href="/apps/{app.name}"
          class="flex items-center gap-4 px-4 sm:px-6 py-4 hover:bg-gray-50 active:bg-gray-100 transition-colors touch-target group"
        >
          <!-- Status Indicator -->
          <div class="flex-shrink-0">
            <span class="block w-3 h-3 rounded-full {statusConfig.color} {app.status === 'deploying' ? 'animate-pulse' : ''}"></span>
          </div>

          <!-- App Info -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-gray-900 truncate group-hover:text-pvd-600 transition-colors">{app.name}</h3>
              {#if app.status !== 'running'}
                <span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {
                  app.status === 'deploying' ? 'bg-yellow-100 text-yellow-700' :
                  app.status === 'failed' ? 'bg-red-100 text-red-700' :
                  'bg-gray-100 text-gray-600'
                }">
                  {statusConfig.label}
                </span>
              {/if}
            </div>
            <p class="text-sm text-gray-500 truncate mt-0.5">
              {#if app.domains?.length}
                {app.domains[0]}
              {:else}
                {app.name}.pvdify.win
              {/if}
            </p>
          </div>

          <!-- Meta & Arrow -->
          <div class="flex items-center gap-4 flex-shrink-0">
            <span class="hidden sm:block text-sm text-gray-400">
              {formatDate(app.updated_at)}
            </span>
            <svg class="w-5 h-5 text-gray-300 group-hover:text-gray-400 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
            </svg>
          </div>
        </a>
      {/each}
    </div>

    <!-- Quick Actions Footer -->
    <div class="mt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm">
      <p class="text-gray-500">
        {apps.length} app{apps.length !== 1 ? 's' : ''} • {apps.filter(a => a.status === 'running').length} running
      </p>
      <a href="https://docs.pvdify.win" target="_blank" rel="noopener noreferrer" class="text-pvd-600 hover:text-pvd-700 inline-flex items-center gap-1">
        View documentation
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
        </svg>
      </a>
    </div>
  {/if}
</div>
