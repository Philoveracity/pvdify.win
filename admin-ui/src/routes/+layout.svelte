<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';

  let mobileMenuOpen = false;

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMobileMenu() {
    mobileMenuOpen = false;
  }

  $: currentPath = $page.url.pathname;
</script>

<div class="min-h-screen flex flex-col">
  <!-- Header -->
  <header class="bg-white border-b border-gray-200 sticky top-0 z-40 safe-top">
    <div class="page-container">
      <div class="flex items-center justify-between h-16">
        <!-- Logo & Desktop Nav -->
        <div class="flex items-center gap-8">
          <a href="/" class="flex items-center gap-2.5 group" on:click={closeMobileMenu}>
            <div class="w-8 h-8 bg-gradient-to-br from-pvd-500 to-pvd-700 rounded-lg flex items-center justify-center shadow-sm group-hover:shadow-md transition-shadow">
              <svg class="h-5 w-5 text-white" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.5l7 3.5-7 3.5L5 8l7-3.5zM4 9.5l7 3.5v6.5l-7-3.5V9.5zm16 0v6.5l-7 3.5V13l7-3.5z"/>
              </svg>
            </div>
            <span class="font-semibold text-lg text-gray-900">Pvdify</span>
          </a>

          <!-- Desktop Navigation -->
          <nav class="hidden md:flex items-center gap-1">
            <a
              href="/"
              class="px-3 py-2 rounded-lg text-sm font-medium transition-colors {currentPath === '/' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'}"
            >
              Dashboard
            </a>
            <a
              href="/apps/new"
              class="px-3 py-2 rounded-lg text-sm font-medium transition-colors {currentPath === '/apps/new' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'}"
            >
              New App
            </a>
          </nav>
        </div>

        <!-- Desktop Actions -->
        <div class="hidden md:flex items-center gap-3">
          <a
            href="https://github.com/Philoveracity/pvdify.win#readme"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-ghost btn-sm gap-1.5"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
            </svg>
            Docs
          </a>
          <a href="/apps/new" class="btn btn-primary btn-sm gap-1.5">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
            </svg>
            Create App
          </a>
        </div>

        <!-- Mobile Menu Button -->
        <button
          on:click|stopPropagation={toggleMobileMenu}
          class="md:hidden btn btn-icon btn-ghost relative z-50"
          aria-label="Toggle menu"
          type="button"
        >
          {#if mobileMenuOpen}
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          {:else}
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
            </svg>
          {/if}
        </button>
      </div>
    </div>

    <!-- Mobile Menu -->
    {#if mobileMenuOpen}
      <div class="md:hidden border-t border-gray-100 bg-white animate-slide-down relative z-50">
        <nav class="page-container py-3 space-y-1">
          <a
            href="/"
            on:click={closeMobileMenu}
            class="flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-colors {currentPath === '/' ? 'bg-pvd-50 text-pvd-700' : 'text-gray-600 hover:bg-gray-50'}"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/>
            </svg>
            Dashboard
          </a>
          <a
            href="/apps/new"
            on:click={closeMobileMenu}
            class="flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-colors {currentPath === '/apps/new' ? 'bg-pvd-50 text-pvd-700' : 'text-gray-600 hover:bg-gray-50'}"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
            </svg>
            Create New App
          </a>
          <a
            href="https://github.com/Philoveracity/pvdify.win#readme"
            target="_blank"
            rel="noopener noreferrer"
            on:click={closeMobileMenu}
            class="flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-50"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
            </svg>
            Documentation
          </a>
        </nav>
      </div>
    {/if}
  </header>

  <!-- Main Content -->
  <main class="flex-1">
    <slot />
  </main>

  <!-- Footer -->
  <footer class="border-t border-gray-200 bg-white mt-auto">
    <div class="page-container py-6">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-gray-500">
        <p>&copy; {new Date().getFullYear()} Pvdify. Heroku-style deployments made simple.</p>
        <div class="flex items-center gap-6">
          <a href="https://github.com/Philoveracity/pvdify.win#readme" class="hover:text-gray-700 transition-colors">Docs</a>
          <a href="https://github.com/Philoveracity/pvdify.win" class="hover:text-gray-700 transition-colors">GitHub</a>
          <a href="/status" class="hover:text-gray-700 transition-colors">Status</a>
        </div>
      </div>
    </div>
  </footer>
</div>

<!-- Mobile overlay when menu is open -->
{#if mobileMenuOpen}
  <button
    class="fixed inset-0 bg-black/20 z-40 md:hidden"
    on:click|stopPropagation={closeMobileMenu}
    aria-label="Close menu"
    type="button"
    style="touch-action: manipulation;"
  />
{/if}
