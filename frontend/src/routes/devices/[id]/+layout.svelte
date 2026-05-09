<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let { children } = $props();

	let id = $derived($page.params.id);
	let currentPath = $derived($page.url.pathname);

	let assetData = $state<any>(null);

	const tabs = $derived([
		{ href: `/devices/${id}`, label: 'Genel Bilgi', exact: true },
		{ href: `/devices/${id}/software`, label: 'Yazılım Envanteri' },
		{ href: `/devices/${id}/forensics`, label: 'Derin Denetim' },
		{ href: `/devices/${id}/commands`, label: 'Komut Merkezi' },
	]);

	function isTabActive(href: string, exact?: boolean): boolean {
		if (exact) return currentPath === href;
		return currentPath.startsWith(href);
	}

	async function fetchAsset() {
		try {
			const res = await fetch(`${API}/assets/${id}`);
			if (res.ok) assetData = await res.json();
		} catch (err) {
			console.error('Cihaz bilgisi çekilemedi:', err);
		}
	}

	onMount(() => {
		fetchAsset();
		const interval = setInterval(fetchAsset, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="device-layout">
	<!-- Cihaz Başlık -->
	<header class="device-header">
		<div class="header-top">
			<a href="/devices" class="back-link">← Cihazlar</a>
			{#if assetData}
				<div class="header-status">
					<span class="status-dot" class:online={assetData.status === 'Online'}></span>
					<span class="status-label" class:online={assetData.status === 'Online'}>{assetData.status}</span>
				</div>
			{/if}
		</div>
		<div class="header-info">
			<h1 class="device-hostname">{assetData?.hostname || '...'}</h1>
			<div class="header-meta">
				<span class="meta-item">ID: <span class="mono">{id}</span></span>
				{#if assetData}
					<span class="meta-divider">|</span>
					<span class="meta-item">Uyum: <span class="compliance" class:good={assetData.compliance_score === 100} class:warn={assetData.compliance_score >= 70 && assetData.compliance_score < 100} class:bad={assetData.compliance_score < 70}>%{assetData.compliance_score ?? 100}</span></span>
				{/if}
			</div>
		</div>

		<!-- Tab Navigasyonu -->
		<nav class="tab-nav">
			{#each tabs as tab}
				<a
					href={tab.href}
					class="tab-item"
					class:active={isTabActive(tab.href, tab.exact)}
				>{tab.label}</a>
			{/each}
		</nav>
	</header>

	<!-- Tab İçeriği -->
	<div class="tab-content">
		{@render children()}
	</div>
</div>

<style>
	.device-layout { color: #e2e8f0; }

	.device-header {
		padding: 20px 32px 0;
		border-bottom: 1px solid #1e2130;
		background: #14161e;
	}

	.header-top {
		display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
	}
	.back-link {
		font-size: 12px; color: #64748b; text-decoration: none; font-weight: 500;
		transition: color 0.15s;
	}
	.back-link:hover { color: #94a3b8; }

	.header-status { display: flex; align-items: center; gap: 6px; }
	.status-dot { width: 8px; height: 8px; border-radius: 50%; background: #ef4444; }
	.status-dot.online { background: #22c55e; }
	.status-label { font-size: 12px; font-weight: 600; color: #ef4444; }
	.status-label.online { color: #22c55e; }

	.header-info { margin-bottom: 16px; }
	.device-hostname { font-size: 22px; font-weight: 700; color: #f1f5f9; margin: 0; font-family: monospace; }
	.header-meta { display: flex; align-items: center; gap: 8px; margin-top: 4px; }
	.meta-item { font-size: 11px; color: #64748b; }
	.meta-divider { color: #2d3348; }
	.mono { font-family: monospace; }
	.compliance { font-weight: 700; }
	.compliance.good { color: #22c55e; }
	.compliance.warn { color: #eab308; }
	.compliance.bad { color: #ef4444; }

	.tab-nav { display: flex; gap: 0; }
	.tab-item {
		padding: 10px 18px;
		font-size: 13px;
		font-weight: 500;
		color: #64748b;
		text-decoration: none;
		border-bottom: 2px solid transparent;
		transition: all 0.15s ease;
	}
	.tab-item:hover { color: #94a3b8; }
	.tab-item.active {
		color: #60a5fa;
		border-bottom-color: #3b82f6;
	}

	.tab-content { padding: 24px 32px; }
</style>
