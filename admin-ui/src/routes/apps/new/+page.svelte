<script lang="ts">
  import { goto } from '$app/navigation';

  let name = '';
  let environment = 'production';
  let loading = false;
  let error: string | null = null;
  let nameError: string | null = null;

  const API_URL = import.meta.env.VITE_API_URL || '';

  function validateName(value: string): string | null {
    if (!value.trim()) return null; // Don't show error for empty
    if (value.length < 3) return 'Name must be at least 3 characters';
    if (value.length > 30) return 'Name must be 30 characters or less';
    if (!/^[a-z]/.test(value)) return 'Must start with a lowercase letter';
    if (!/^[a-z][a-z0-9-]*$/.test(value)) return 'Only lowercase letters, numbers, and hyphens allowed';
    if (value.endsWith('-')) return 'Cannot end with a hyphen';
    return null;
  }

  $: nameError = validateName(name);
  $: isValid = name.trim() && !nameError;

  async function createApp() {
    if (!isValid) {
      error = nameError || 'Please enter a valid app name';
      return;
    }

    loading = true;
    error = null;

    try {
      const res = await fetch(`${API_URL}/api/v1/apps`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), environment })
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to create app');
      }

      const app = await res.json();
      goto(`/apps/${app.name}`);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error';
    } finally {
      loading = false;
    }
  }

  function handleSubmit(e: Event) {
    e.preventDefault();
    createApp();
  }

  function handleNameInput(e: Event) {
    const input = e.target as HTMLInputElement;
    // Auto-convert to lowercase and replace spaces/underscores with hyphens
    name = input.value.toLowerCase().replace(/[\s_]+/g, '-');
  }
</script>

<svelte:head>
  <title>Create App | Pvdify</title>
</svelte:head>

<div class="page-container py-6 sm:py-8">
  <!-- Breadcrumb -->
  <nav class="mb-6">
    <a href="/" class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 transition-colors touch-target -ml-2 pl-2">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
      </svg>
      Back to Dashboard
    </a>
  </nav>

  <div class="max-w-2xl mx-auto">
    <!-- Main Form Card -->
    <div class="card">
      <div class="px-4 sm:px-6 py-5 sm:py-6 border-b border-gray-200">
        <h1 class="text-xl font-semibold text-gray-900">Create New App</h1>
        <p class="mt-1 text-sm text-gray-500">Deploy containerized applications with zero configuration</p>
      </div>

      <form on:submit={handleSubmit}>
        <div class="px-4 sm:px-6 py-5 sm:py-6 space-y-6">
          <!-- Error Alert -->
          {#if error}
            <div class="alert alert-error">
              <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <span>{error}</span>
            </div>
          {/if}

          <!-- App Name Field -->
          <div class="form-group">
            <label for="name" class="form-label">
              App Name
              <span class="text-red-500">*</span>
            </label>
            <div class="mt-1.5 relative">
              <input
                type="text"
                id="name"
                bind:value={name}
                on:input={handleNameInput}
                class="input {nameError && name ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : ''}"
                placeholder="my-awesome-app"
                autocomplete="off"
                autocapitalize="off"
                spellcheck="false"
                maxlength="30"
              />
              {#if name && !nameError}
                <div class="absolute right-3 top-1/2 -translate-y-1/2">
                  <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                </div>
              {/if}
            </div>
            {#if nameError && name}
              <p class="mt-1.5 text-sm text-red-600">{nameError}</p>
            {:else}
              <p class="mt-1.5 text-xs text-gray-500">
                Lowercase letters, numbers, and hyphens only. Will be available at <span class="font-medium">{name || 'your-app'}.pvdify.win</span>
              </p>
            {/if}
          </div>

          <!-- Environment Field -->
          <div class="form-group">
            <label for="environment" class="form-label">Environment</label>
            <div class="mt-1.5">
              <div class="grid grid-cols-2 gap-3">
                <label class="relative flex cursor-pointer rounded-lg border p-4 focus:outline-none {environment === 'production' ? 'border-pvd-500 ring-2 ring-pvd-500 bg-pvd-50' : 'border-gray-200 hover:border-gray-300'}">
                  <input
                    type="radio"
                    name="environment"
                    value="production"
                    bind:group={environment}
                    class="sr-only"
                  />
                  <div class="flex flex-col">
                    <span class="block text-sm font-medium {environment === 'production' ? 'text-pvd-900' : 'text-gray-900'}">
                      Production
                    </span>
                    <span class="mt-1 text-xs {environment === 'production' ? 'text-pvd-700' : 'text-gray-500'}">
                      Live traffic
                    </span>
                  </div>
                  {#if environment === 'production'}
                    <svg class="w-5 h-5 text-pvd-600 absolute top-3 right-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                    </svg>
                  {/if}
                </label>

                <label class="relative flex cursor-pointer rounded-lg border p-4 focus:outline-none {environment === 'staging' ? 'border-pvd-500 ring-2 ring-pvd-500 bg-pvd-50' : 'border-gray-200 hover:border-gray-300'}">
                  <input
                    type="radio"
                    name="environment"
                    value="staging"
                    bind:group={environment}
                    class="sr-only"
                  />
                  <div class="flex flex-col">
                    <span class="block text-sm font-medium {environment === 'staging' ? 'text-pvd-900' : 'text-gray-900'}">
                      Staging
                    </span>
                    <span class="mt-1 text-xs {environment === 'staging' ? 'text-pvd-700' : 'text-gray-500'}">
                      Testing only
                    </span>
                  </div>
                  {#if environment === 'staging'}
                    <svg class="w-5 h-5 text-pvd-600 absolute top-3 right-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                    </svg>
                  {/if}
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- Form Actions -->
        <div class="px-4 sm:px-6 py-4 bg-gray-50 border-t border-gray-200 flex flex-col-reverse sm:flex-row items-stretch sm:items-center justify-end gap-3">
          <a href="/" class="btn btn-secondary justify-center">Cancel</a>
          <button
            type="submit"
            class="btn btn-primary justify-center gap-2"
            disabled={loading || !isValid}
          >
            {#if loading}
              <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Creating...
            {:else}
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
              </svg>
              Create App
            {/if}
          </button>
        </div>
      </form>
    </div>

    <!-- Next Steps Card -->
    <div class="card mt-6">
      <div class="px-4 sm:px-6 py-4 border-b border-gray-200">
        <h2 class="font-medium text-gray-900">What happens next?</h2>
      </div>
      <div class="px-4 sm:px-6 py-4 sm:py-6">
        <ol class="space-y-4">
          <li class="flex gap-4">
            <span class="flex-shrink-0 w-7 h-7 rounded-full bg-pvd-100 text-pvd-700 flex items-center justify-center text-sm font-medium">1</span>
            <div class="pt-0.5">
              <div class="font-medium text-gray-900">Configure environment</div>
              <div class="text-sm text-gray-500 mt-0.5">Set config vars like <code class="bg-gray-100 px-1.5 py-0.5 rounded text-xs">DATABASE_URL</code> and <code class="bg-gray-100 px-1.5 py-0.5 rounded text-xs">API_KEY</code></div>
            </div>
          </li>
          <li class="flex gap-4">
            <span class="flex-shrink-0 w-7 h-7 rounded-full bg-pvd-100 text-pvd-700 flex items-center justify-center text-sm font-medium">2</span>
            <div class="pt-0.5">
              <div class="font-medium text-gray-900">Deploy your container</div>
              <div class="text-sm text-gray-500 mt-0.5">
                <code class="bg-gray-900 text-gray-100 px-2 py-1 rounded text-xs inline-block">pvdify deploy APP --image IMAGE</code>
              </div>
            </div>
          </li>
          <li class="flex gap-4">
            <span class="flex-shrink-0 w-7 h-7 rounded-full bg-pvd-100 text-pvd-700 flex items-center justify-center text-sm font-medium">3</span>
            <div class="pt-0.5">
              <div class="font-medium text-gray-900">Add a custom domain</div>
              <div class="text-sm text-gray-500 mt-0.5">Point your domain to your app with a CNAME record</div>
            </div>
          </li>
        </ol>
      </div>
    </div>

    <!-- CLI Hint -->
    <div class="mt-6 text-center">
      <p class="text-sm text-gray-500">
        Prefer the command line?
        <code class="ml-1 bg-gray-100 px-2 py-1 rounded text-xs text-gray-700">pvdify apps:create {name || 'my-app'}</code>
      </p>
    </div>
  </div>
</div>
