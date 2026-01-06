<script lang="ts">
  import { onMount } from 'svelte';

  interface ServiceStatus {
    name: string;
    status: 'operational' | 'degraded' | 'down' | 'checking';
    latency?: number;
    message?: string;
  }

  let services: ServiceStatus[] = [
    { name: 'API', status: 'checking' },
    { name: 'Database', status: 'checking' },
    { name: 'Container Runtime', status: 'checking' }
  ];

  let lastChecked: Date | null = null;
  let overallStatus: 'operational' | 'degraded' | 'down' = 'operational';

  const API_URL = import.meta.env.VITE_API_URL || '';

  onMount(() => {
    checkStatus();
    const interval = setInterval(checkStatus, 30000);
    return () => clearInterval(interval);
  });

  async function checkStatus() {
    // Check API
    const apiStart = performance.now();
    try {
      const res = await fetch(`${API_URL}/api/v1/apps`);
      const apiLatency = Math.round(performance.now() - apiStart);
      if (res.ok) {
        services[0] = { name: 'API', status: 'operational', latency: apiLatency };
        services[1] = { name: 'Database', status: 'operational' };
        services[2] = { name: 'Container Runtime', status: 'operational' };
      } else {
        services[0] = { name: 'API', status: 'degraded', message: `HTTP ${res.status}` };
      }
    } catch (e) {
      services[0] = { name: 'API', status: 'down', message: 'Connection failed' };
      services[1] = { name: 'Database', status: 'down' };
      services[2] = { name: 'Container Runtime', status: 'down' };
    }

    // Update overall status
    if (services.some(s => s.status === 'down')) {
      overallStatus = 'down';
    } else if (services.some(s => s.status === 'degraded')) {
      overallStatus = 'degraded';
    } else {
      overallStatus = 'operational';
    }

    lastChecked = new Date();
    services = services;
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'operational': return 'bg-green-500';
      case 'degraded': return 'bg-yellow-500';
      case 'down': return 'bg-red-500';
      default: return 'bg-gray-400 animate-pulse';
    }
  }

  function getStatusText(status: string) {
    switch (status) {
      case 'operational': return 'Operational';
      case 'degraded': return 'Degraded';
      case 'down': return 'Down';
      default: return 'Checking...';
    }
  }

  function getOverallBg(status: string) {
    switch (status) {
      case 'operational': return 'bg-green-50 border-green-200';
      case 'degraded': return 'bg-yellow-50 border-yellow-200';
      case 'down': return 'bg-red-50 border-red-200';
      default: return 'bg-gray-50 border-gray-200';
    }
  }

  function getOverallText(status: string) {
    switch (status) {
      case 'operational': return 'text-green-700';
      case 'degraded': return 'text-yellow-700';
      case 'down': return 'text-red-700';
      default: return 'text-gray-700';
    }
  }
</script>

<svelte:head>
  <title>Status | Pvdify</title>
</svelte:head>

<div class="page-container py-6 sm:py-8">
  <nav class="mb-6">
    <a href="/" class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 transition-colors touch-target -ml-2 pl-2">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
      </svg>
      Back to Dashboard
    </a>
  </nav>

  <div class="max-w-2xl mx-auto">
    <div class="text-center mb-8">
      <h1 class="text-2xl font-semibold text-gray-900">System Status</h1>
      <p class="mt-2 text-gray-500">Current status of Pvdify services</p>
    </div>

    <!-- Overall Status Banner -->
    <div class="card {getOverallBg(overallStatus)} mb-6">
      <div class="px-6 py-8 text-center">
        <div class="inline-flex items-center gap-3">
          <span class="w-4 h-4 rounded-full {getStatusColor(overallStatus)}"></span>
          <span class="text-xl font-semibold {getOverallText(overallStatus)}">
            {#if overallStatus === 'operational'}
              All Systems Operational
            {:else if overallStatus === 'degraded'}
              Some Systems Degraded
            {:else}
              System Outage
            {/if}
          </span>
        </div>
      </div>
    </div>

    <!-- Individual Services -->
    <div class="card">
      <div class="card-header">
        <h2 class="font-medium text-gray-900">Services</h2>
        {#if lastChecked}
          <span class="text-sm text-gray-500">
            Updated {lastChecked.toLocaleTimeString()}
          </span>
        {/if}
      </div>
      <div class="divide-y divide-gray-100">
        {#each services as service}
          <div class="px-4 sm:px-6 py-4 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <span class="w-3 h-3 rounded-full {getStatusColor(service.status)}"></span>
              <span class="font-medium text-gray-900">{service.name}</span>
            </div>
            <div class="flex items-center gap-3">
              {#if service.latency}
                <span class="text-sm text-gray-400">{service.latency}ms</span>
              {/if}
              {#if service.message}
                <span class="text-sm text-gray-500">{service.message}</span>
              {/if}
              <span class="text-sm {service.status === 'operational' ? 'text-green-600' : service.status === 'degraded' ? 'text-yellow-600' : service.status === 'down' ? 'text-red-600' : 'text-gray-500'}">
                {getStatusText(service.status)}
              </span>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Refresh Button -->
    <div class="mt-6 text-center">
      <button on:click={checkStatus} class="btn btn-secondary btn-sm gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
        </svg>
        Refresh
      </button>
    </div>
  </div>
</div>
