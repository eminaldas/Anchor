<script lang="ts">
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';

	let dashboardData = $state<any>(null);
	let loading = $state(true);

	async function fetchDashboard() {
		try {
			const res = await fetch(`${API}/forensics/dashboard`);
			if (res.ok) dashboardData = await res.json();
		} catch (err) {
			console.error('Dashboard verisi çekilemedi:', err);
		} finally {
			loading = false;
		}
	}

	function severityColor(s: string) {
		return { Critical: '#ef4444', High: '#f97316', Medium: '#eab308', Low: '#3b82f6', Info: '#6b7280' }[s] || '#6b7280';
	}

	function severityBg(s: string) {
		return { Critical: 'rgba(239,68,68,0.1)', High: 'rgba(249,115,22,0.1)', Medium: 'rgba(234,179,8,0.1)', Low: 'rgba(59,130,246,0.1)' }[s] || 'rgba(107,114,128,0.1)';
	}

	function eventTypeLabel(t: string) {
		return {
			logon_success: 'Başarılı Giriş',
			logon_failed: 'Başarısız Giriş',
			logoff: 'Oturum Kapatma',
			user_created: 'Kullanıcı Oluşturma',
			group_member_added: 'Gruba Ekleme'
		}[t] || t;
	}

	function eventTypeColor(t: string) {
		return {
			logon_success: '#22c55e',
			logon_failed: '#ef4444',
			logoff: '#6b7280',
			user_created: '#f97316',
			group_member_added: '#f97316'
		}[t] || '#6b7280';
	}

	function formatTime(d: string) {
		return new Date(d).toLocaleString('tr-TR', {
			day: '2-digit', month: '2-digit',
			hour: '2-digit', minute: '2-digit', second: '2-digit'
		});
	}

	function checkTypeLabel(ct: string) {
		const labels: Record<string, string> = {
			'Policy_Violation': 'Politika İhlali',
			'Defender_Check': 'Defender Kontrolü',
			'Uptime_Check': 'Uptime Kontrolü',
			'Deep_Admin_Violation': 'Yetkisiz Admin',
			'Deep_BruteForce': 'Brute Force',
			'Deep_UserCreated': 'Yeni Kullanıcı',
			'Deep_GroupChange': 'Grup Değişikliği',
			'Deep_AfterHoursLogin': 'Mesai Dışı Giriş',
			'Deep_FIM_Violation': 'Dosya Bütünlüğü İhlali',
		};
		return labels[ct] || ct;
	}

	onMount(() => {
		fetchDashboard();
		const interval = setInterval(fetchDashboard, 10000);
		return () => clearInterval(interval);
	});
</script>

<div class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Tehdit Merkezi</h1>
			<p class="page-subtitle">Gerçek zamanlı güvenlik olayları ve bulgu akışı</p>
		</div>
		<div class="header-badge">
			<span class="pulse-dot"></span>
			<span>Canlı İzleme</span>
		</div>
	</header>

	{#if loading}
		<div class="loading-state">Veriler yükleniyor...</div>
	{:else if dashboardData}
		<!-- Özet Kartları -->
		<div class="stats-grid">
			<div class="stat-card">
				<span class="stat-label">Açık Bulgular</span>
				<span class="stat-value">{dashboardData.stats.total_open_findings}</span>
				<div class="stat-breakdown">
					<span style="color: #ef4444">{dashboardData.stats.critical_count} Kritik</span>
					<span style="color: #f97316">{dashboardData.stats.high_count} Yüksek</span>
					<span style="color: #eab308">{dashboardData.stats.medium_count} Orta</span>
				</div>
			</div>
			<div class="stat-card">
				<span class="stat-label">Aktif Cihaz</span>
				<span class="stat-value">{dashboardData.stats.online_assets} <span class="stat-sub">/ {dashboardData.stats.total_assets}</span></span>
			</div>
			<div class="stat-card">
				<span class="stat-label">Başarısız Giriş</span>
				<span class="stat-value" style="color: #ef4444">{dashboardData.stats.failed_logins}</span>
			</div>
			<div class="stat-card">
				<span class="stat-label">FIM İhlali</span>
				<span class="stat-value" style="color: #f97316">{dashboardData.stats.fim_changes}</span>
			</div>
		</div>

		<div class="content-grid">
			<!-- Bulgu Feed'i -->
			<div class="panel">
				<div class="panel-header">
					<h2 class="panel-title">Son Bulgular</h2>
					<span class="panel-count">{dashboardData.recent_findings?.length || 0}</span>
				</div>
				<div class="panel-body scrollable">
					{#if dashboardData.recent_findings?.length > 0}
						{#each dashboardData.recent_findings as finding}
							<div class="feed-item" style="border-left-color: {severityColor(finding.severity)}">
								<div class="feed-row">
									<span class="severity-badge" style="background: {severityBg(finding.severity)}; color: {severityColor(finding.severity)}">{finding.severity}</span>
									<span class="feed-type">{checkTypeLabel(finding.check_type)}</span>
									<span class="feed-time">{formatTime(finding.created_at)}</span>
								</div>
								<p class="feed-desc">{finding.description}</p>
								<span class="feed-asset">{finding.asset_id.split('_')[0]}</span>
							</div>
						{/each}
					{:else}
						<div class="empty-state">Açık bulgu bulunmuyor.</div>
					{/if}
				</div>
			</div>

			<!-- Son Güvenlik Olayları -->
			<div class="panel">
				<div class="panel-header">
					<h2 class="panel-title">Son Güvenlik Olayları</h2>
					<span class="panel-count">{dashboardData.recent_events?.length || 0}</span>
				</div>
				<div class="panel-body scrollable">
					{#if dashboardData.recent_events?.length > 0}
						<table class="data-table">
							<thead>
								<tr>
									<th>Zaman</th>
									<th>Olay</th>
									<th>Kullanıcı</th>
									<th>IP</th>
								</tr>
							</thead>
							<tbody>
								{#each dashboardData.recent_events as evt}
									<tr>
										<td class="mono">{formatTime(evt.event_time)}</td>
										<td><span class="event-badge" style="color: {eventTypeColor(evt.event_type)}">{eventTypeLabel(evt.event_type)}</span></td>
										<td class="mono">{evt.target_user || '-'}</td>
										<td class="mono">{evt.source_ip || '-'}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{:else}
						<div class="empty-state">Güvenlik olayı bulunmuyor.</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- FIM Alarmları -->
		{#if dashboardData.fim_alerts?.length > 0}
		<div class="panel" style="margin-top: 20px;">
			<div class="panel-header">
				<h2 class="panel-title" style="color: #f97316">⚠ Dosya Bütünlüğü İhlalleri</h2>
			</div>
			<div class="panel-body">
				<table class="data-table">
					<thead>
						<tr>
							<th>Dosya Yolu</th>
							<th>Önceki Hash</th>
							<th>Yeni Hash</th>
							<th>Tarih</th>
						</tr>
					</thead>
					<tbody>
						{#each dashboardData.fim_alerts as alert}
							<tr class="fim-alert-row">
								<td class="mono">{alert.file_path}</td>
								<td class="mono hash-cell">{alert.previous_hash ? alert.previous_hash.slice(0, 12) + '...' : '-'}</td>
								<td class="mono hash-cell">{alert.sha256.slice(0, 12)}...</td>
								<td class="mono">{formatTime(alert.created_at)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
		{/if}
	{/if}
</div>

<style>
	.page { padding: 28px 32px; color: #e2e8f0; }

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 24px;
		padding-bottom: 16px;
		border-bottom: 1px solid #1e2130;
	}
	.page-title { font-size: 22px; font-weight: 700; color: #f1f5f9; margin: 0; }
	.page-subtitle { font-size: 13px; color: #64748b; margin-top: 4px; }

	.header-badge {
		display: flex; align-items: center; gap: 8px;
		background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3);
		padding: 6px 14px; border-radius: 6px; font-size: 12px; color: #ef4444; font-weight: 600;
	}
	.pulse-dot {
		width: 8px; height: 8px; border-radius: 50%; background: #ef4444;
		animation: pulse 2s infinite;
	}
	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
		margin-bottom: 24px;
	}
	.stat-card {
		background: #14161e;
		border: 1px solid #1e2130;
		border-radius: 8px;
		padding: 18px 20px;
	}
	.stat-label { font-size: 11px; color: #64748b; text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600; display: block; }
	.stat-value { font-size: 28px; font-weight: 700; color: #f1f5f9; display: block; margin-top: 4px; font-family: monospace; }
	.stat-sub { font-size: 16px; color: #64748b; font-weight: 400; }
	.stat-breakdown { display: flex; gap: 12px; margin-top: 8px; font-size: 11px; font-weight: 600; }

	.content-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 20px;
	}

	.panel {
		background: #14161e;
		border: 1px solid #1e2130;
		border-radius: 8px;
		overflow: hidden;
	}
	.panel-header {
		display: flex; justify-content: space-between; align-items: center;
		padding: 14px 18px; border-bottom: 1px solid #1e2130;
	}
	.panel-title { font-size: 14px; font-weight: 600; color: #cbd5e1; margin: 0; }
	.panel-count {
		background: #1e293b; color: #64748b; font-size: 11px; font-weight: 600;
		padding: 2px 8px; border-radius: 10px;
	}
	.panel-body { padding: 0; }
	.scrollable { max-height: 420px; overflow-y: auto; }

	.feed-item {
		padding: 12px 18px;
		border-bottom: 1px solid #1a1d2a;
		border-left: 3px solid transparent;
	}
	.feed-item:hover { background: #181b27; }
	.feed-row { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
	.severity-badge {
		font-size: 10px; font-weight: 700; text-transform: uppercase;
		padding: 2px 8px; border-radius: 4px; letter-spacing: 0.5px;
	}
	.feed-type { font-size: 12px; color: #94a3b8; font-weight: 500; }
	.feed-time { font-size: 11px; color: #475569; margin-left: auto; font-family: monospace; }
	.feed-desc { font-size: 12px; color: #cbd5e1; margin: 2px 0 4px 0; line-height: 1.4; }
	.feed-asset { font-size: 10px; color: #475569; font-family: monospace; }

	.data-table { width: 100%; font-size: 12px; border-collapse: collapse; }
	.data-table thead { background: #0f1117; position: sticky; top: 0; }
	.data-table th {
		text-align: left; padding: 8px 14px; color: #64748b;
		font-weight: 600; text-transform: uppercase; font-size: 10px; letter-spacing: 0.5px;
		border-bottom: 1px solid #1e2130;
	}
	.data-table td { padding: 8px 14px; border-bottom: 1px solid #1a1d2a; color: #cbd5e1; }
	.data-table tbody tr:hover { background: #181b27; }
	.mono { font-family: monospace; font-size: 11px; }
	.event-badge { font-weight: 600; font-size: 11px; }
	.hash-cell { font-size: 10px; color: #64748b; }

	.fim-alert-row { background: rgba(249, 115, 22, 0.05); }

	.empty-state { padding: 32px; text-align: center; color: #475569; font-size: 13px; }
	.loading-state { padding: 64px; text-align: center; color: #475569; font-size: 14px; }

	.scrollable::-webkit-scrollbar { width: 5px; }
	.scrollable::-webkit-scrollbar-track { background: transparent; }
	.scrollable::-webkit-scrollbar-thumb { background: #2d3348; border-radius: 4px; }
</style>