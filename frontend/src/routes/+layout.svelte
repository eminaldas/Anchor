<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/stores';

	let { children } = $props();

	let currentPath = $derived($page.url.pathname);

	const navItems = [
		{ href: '/', label: 'Tehdit Merkezi', icon: '🔴' },
		{ href: '/devices', label: 'Cihazlar', icon: '💻' },
		{ href: '/admin/policies', label: 'Politikalar', icon: '🛡️' },
		{ href: '/admin/whitelist', label: 'Admin Whitelist', icon: '👤' },
		{ href: '/admin/fim-config', label: 'FIM Yapılandırma', icon: '📁' },
	];

	function isActive(href: string): boolean {
		if (href === '/') return currentPath === '/';
		return currentPath.startsWith(href);
	}
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="app-shell">
	<aside class="sidebar">
		<div class="sidebar-brand">
			<span class="brand-icon">⬡</span>
			<span class="brand-text">Sentinel GRC</span>
		</div>
		<nav class="sidebar-nav">
			{#each navItems as item}
				<a
					href={item.href}
					class="nav-item"
					class:active={isActive(item.href)}
				>
					<span class="nav-icon">{item.icon}</span>
					<span class="nav-label">{item.label}</span>
				</a>
			{/each}
		</nav>
		<div class="sidebar-footer">
			<span class="version-tag">v1.2.0</span>
		</div>
	</aside>
	<main class="main-content">
		{@render children()}
	</main>
</div>

<style>
	.app-shell {
		display: flex;
		min-height: 100vh;
		background: #0f1117;
	}

	.sidebar {
		width: 220px;
		background: #14161e;
		border-right: 1px solid #1e2130;
		display: flex;
		flex-direction: column;
		position: fixed;
		top: 0;
		left: 0;
		bottom: 0;
		z-index: 40;
	}

	.sidebar-brand {
		padding: 20px 16px;
		border-bottom: 1px solid #1e2130;
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.brand-icon {
		font-size: 22px;
		color: #3b82f6;
	}

	.brand-text {
		font-size: 15px;
		font-weight: 700;
		color: #e2e8f0;
		letter-spacing: 0.5px;
	}

	.sidebar-nav {
		padding: 12px 8px;
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 6px;
		color: #8891a5;
		text-decoration: none;
		font-size: 13px;
		font-weight: 500;
		transition: all 0.15s ease;
	}

	.nav-item:hover {
		background: #1a1d2e;
		color: #c8cedd;
	}

	.nav-item.active {
		background: #1e293b;
		color: #60a5fa;
	}

	.nav-icon {
		font-size: 16px;
		width: 22px;
		text-align: center;
	}

	.sidebar-footer {
		padding: 12px 16px;
		border-top: 1px solid #1e2130;
	}

	.version-tag {
		font-size: 10px;
		color: #4a5568;
		font-family: monospace;
	}

	.main-content {
		flex: 1;
		margin-left: 220px;
		min-height: 100vh;
	}
</style>
