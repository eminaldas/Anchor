<script lang="ts">
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let assets = $state<any[]>([]);
	let loading = $state(true);

	async function fetchAssets() {
		try {
			const res = await fetch(`${API}/assets`);
			if (res.ok) assets = await res.json() || [];
		} catch (err) {
			console.error('Cihazlar çekilemedi:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchAssets();
		const interval = setInterval(fetchAssets, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Kayıtlı Cihazlar</h1>
			<p class="page-subtitle">Sentinel ajanları tarafından izlenen tüm cihazlar</p>
		</div>
		<div class="device-count">{assets.length} Cihaz</div>
	</header>

	{#if loading}
		<div class="loading-state">Cihazlar yükleniyor...</div>
	{:else if assets.length === 0}
		<div class="empty-state">
			<p>Henüz kayıtlı cihaz bulunmuyor.</p>
			<p class="empty-hint">Ajanların bağlantı kurması bekleniyor...</p>
		</div>
	{:else}
		<div class="device-grid">
			{#each assets as asset}
				<a href="/devices/{asset.id}" class="device-card">
					<div class="card-header">
						<span class="hostname">{asset.hostname}</span>
						<span class="status-dot" class:online={asset.status === 'Online'}></span>
					</div>
					<div class="card-id">{asset.id}</div>
					<div class="card-metrics">
						<div class="metric">
							<span class="metric-label">CPU</span>
							<span class="metric-value">{asset.cpu_usage || 'N/A'}</span>
						</div>
						<div class="metric">
							<span class="metric-label">RAM</span>
							<span class="metric-value">{asset.ram_usage || 'N/A'}</span>
						</div>
						<div class="metric">
							<span class="metric-label">Uyum</span>
							<span class="metric-value" class:score-good={asset.compliance_score === 100} class:score-warn={asset.compliance_score >= 70 && asset.compliance_score < 100} class:score-bad={asset.compliance_score < 70}>
								%{asset.compliance_score ?? 100}
							</span>
						</div>
					</div>
					<div class="card-footer">
						<span class="status-text" class:online={asset.status === 'Online'}>{asset.status}</span>
						<span class="last-seen">Son: {new Date(asset.last_seen).toLocaleString('tr-TR', { hour: '2-digit', minute: '2-digit' })}</span>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>

<style>
	.page { padding: 28px 32px; color: #e2e8f0; }
	.page-header {
		display: flex; justify-content: space-between; align-items: flex-start;
		margin-bottom: 24px; padding-bottom: 16px; border-bottom: 1px solid #1e2130;
	}
	.page-title { font-size: 22px; font-weight: 700; color: #f1f5f9; margin: 0; }
	.page-subtitle { font-size: 13px; color: #64748b; margin-top: 4px; }
	.device-count {
		background: #1e293b; color: #60a5fa; font-size: 13px; font-weight: 600;
		padding: 6px 16px; border-radius: 6px; border: 1px solid #2d3a50;
	}

	.device-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 16px;
	}

	.device-card {
		background: #14161e;
		border: 1px solid #1e2130;
		border-radius: 8px;
		padding: 18px 20px;
		text-decoration: none;
		color: inherit;
		transition: border-color 0.15s ease, box-shadow 0.15s ease;
	}
	.device-card:hover {
		border-color: #3b82f6;
		box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2);
	}

	.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
	.hostname { font-size: 15px; font-weight: 600; color: #60a5fa; font-family: monospace; }
	.status-dot { width: 8px; height: 8px; border-radius: 50%; background: #ef4444; }
	.status-dot.online { background: #22c55e; }
	.card-id { font-size: 10px; color: #475569; font-family: monospace; margin-bottom: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	.card-metrics { display: flex; gap: 8px; margin-bottom: 14px; }
	.metric {
		flex: 1; background: #0f1117; border: 1px solid #1e2130;
		border-radius: 6px; padding: 8px 10px;
	}
	.metric-label { display: block; font-size: 10px; color: #64748b; text-transform: uppercase; font-weight: 600; }
	.metric-value { display: block; font-size: 13px; color: #e2e8f0; font-family: monospace; font-weight: 600; margin-top: 2px; }
	.score-good { color: #22c55e; }
	.score-warn { color: #eab308; }
	.score-bad { color: #ef4444; }

	.card-footer { display: flex; justify-content: space-between; align-items: center; }
	.status-text { font-size: 12px; font-weight: 600; color: #ef4444; }
	.status-text.online { color: #22c55e; }
	.last-seen { font-size: 10px; color: #475569; font-family: monospace; }

	.empty-state { padding: 64px; text-align: center; color: #64748b; background: #14161e; border: 1px solid #1e2130; border-radius: 8px; }
	.empty-hint { font-size: 12px; color: #475569; margin-top: 4px; }
	.loading-state { padding: 64px; text-align: center; color: #475569; }
</style>
