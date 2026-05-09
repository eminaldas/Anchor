<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let id = $derived($page.params.id);

	let summary = $state<any>(null);
	let adminSnapshot = $state<any>(null);
	let events = $state<any[]>([]);
	let fimRecords = $state<any[]>([]);
	let whitelist = $state<any[]>([]);

	let eventFilter = $state('');
	let openSection = $state('admins'); // admins, events, fim

	let whitelistMap = $derived.by(() => {
		const m: Record<string, boolean> = {};
		for (const w of whitelist) m[w.username.toLowerCase()] = true;
		return m;
	});

	let adminList = $derived.by(() => {
		if (!adminSnapshot?.data) return [];
		try {
			const data = typeof adminSnapshot.data === 'string' ? JSON.parse(adminSnapshot.data) : adminSnapshot.data;
			return Array.isArray(data) ? data : [];
		} catch { return []; }
	});

	let filteredEvents = $derived.by(() => {
		if (!eventFilter) return events;
		return events.filter(e => e.event_type === eventFilter);
	});

	let fimChangedCount = $derived(fimRecords.filter(r => r.status === 'changed').length);

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
		return { logon_success: '#22c55e', logon_failed: '#ef4444', logoff: '#6b7280', user_created: '#f97316', group_member_added: '#f97316' }[t] || '#6b7280';
	}

	function logonTypeLabel(lt: number) {
		return { 2: 'Interactive', 3: 'Network', 4: 'Batch', 5: 'Service', 7: 'Unlock', 10: 'RemoteDesktop', 11: 'CachedInteractive' }[lt] || `Type ${lt}`;
	}

	function fimStatusLabel(s: string) {
		return { baseline: 'Baseline', unchanged: 'Değişmedi', changed: 'Değişti' }[s] || s;
	}

	function fimStatusColor(s: string) {
		return { baseline: '#3b82f6', unchanged: '#22c55e', changed: '#ef4444' }[s] || '#6b7280';
	}

	function shortHash(h: string) {
		if (!h || h.length < 16) return h || '-';
		return h.slice(0, 8) + '...' + h.slice(-8);
	}

	function formatTime(d: string) {
		return new Date(d).toLocaleString('tr-TR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	function isUnauthorized(admin: string): boolean {
		const lower = admin.toLowerCase();
		const parts = lower.split('\\');
		const username = parts.length > 1 ? parts[parts.length - 1] : lower;
		return !whitelistMap[username] && !whitelistMap[lower];
	}

	let unauthorizedCount = $derived(adminList.filter((a: string) => isUnauthorized(a)).length);

	async function fetchAll() {
		try {
			const [sumRes, snapRes, evtRes, fimRes, wlRes] = await Promise.all([
				fetch(`${API}/forensics/summary/${id}`),
				fetch(`${API}/forensics/snapshots?asset_id=${id}&type=local_admins`),
				fetch(`${API}/forensics/events?asset_id=${id}`),
				fetch(`${API}/forensics/fim?asset_id=${id}`),
				fetch(`${API}/admin/whitelist`)
			]);
			if (sumRes.ok) summary = await sumRes.json();
			if (snapRes.ok) {
				const snaps = await snapRes.json();
				adminSnapshot = snaps?.length > 0 ? snaps[0] : null;
			}
			if (evtRes.ok) events = await evtRes.json() || [];
			if (fimRes.ok) fimRecords = await fimRes.json() || [];
			if (wlRes.ok) whitelist = await wlRes.json() || [];
		} catch (err) { console.error(err); }
	}

	onMount(() => {
		fetchAll();
		const interval = setInterval(fetchAll, 15000);
		return () => clearInterval(interval);
	});
</script>

<div>
	<!-- Forensics Sağlık Özeti -->
	{#if summary}
	<div class="health-bar">
		<div class="health-item" class:healthy={unauthorizedCount === 0} class:alert={unauthorizedCount > 0}>
			<span class="health-dot"></span>
			<span>Admin Denetimi</span>
		</div>
		<div class="health-item" class:healthy={summary.event_stats?.failed_logins === 0} class:alert={summary.event_stats?.failed_logins > 0}>
			<span class="health-dot"></span>
			<span>Olay Kaydı</span>
		</div>
		<div class="health-item" class:healthy={fimChangedCount === 0} class:alert={fimChangedCount > 0}>
			<span class="health-dot"></span>
			<span>Dosya Bütünlüğü</span>
		</div>
		<span class="health-summary">
			{#if unauthorizedCount === 0 && (summary.event_stats?.failed_logins || 0) === 0 && fimChangedCount === 0}
				Tüm modüller sağlıklı
			{:else}
				{(unauthorizedCount > 0 ? 1 : 0) + ((summary.event_stats?.failed_logins || 0) > 0 ? 1 : 0) + (fimChangedCount > 0 ? 1 : 0)} modülde alarm
			{/if}
		</span>
	</div>
	{/if}

	<!-- Bölüm Seçici -->
	<div class="section-tabs">
		<button class="section-tab" class:active={openSection === 'admins'} onclick={() => openSection = 'admins'}>🔐 Yerel Admin Denetimi</button>
		<button class="section-tab" class:active={openSection === 'events'} onclick={() => openSection = 'events'}>📋 Güvenlik Olayları</button>
		<button class="section-tab" class:active={openSection === 'fim'} onclick={() => openSection = 'fim'}>🔒 Dosya Bütünlüğü</button>
	</div>

	<!-- PANEL 1: Yerel Admin Denetimi -->
	{#if openSection === 'admins'}
	<div class="panel">
		<div class="panel-header">
			<h3 class="panel-title">Yerel Administrator Hesapları</h3>
			{#if adminSnapshot}
				<span class="panel-meta">Son tarama: {formatTime(adminSnapshot.created_at)}</span>
			{/if}
		</div>
		<div class="panel-body">
			{#if adminList.length === 0}
				<div class="empty-state">Admin denetim verisi henüz alınmadı.</div>
			{:else}
				{#if unauthorizedCount > 0}
					<div class="alert-banner alert-red">
						⚠ {unauthorizedCount} yetkisiz administrator tespit edildi!
					</div>
				{/if}
				<table class="data-table">
					<thead>
						<tr>
							<th>Kullanıcı Adı</th>
							<th>Yetki Durumu</th>
						</tr>
					</thead>
					<tbody>
						{#each adminList as admin}
							<tr class:unauthorized-row={isUnauthorized(admin)}>
								<td class="mono">{admin}</td>
								<td>
									{#if isUnauthorized(admin)}
										<span class="status-tag tag-red">❌ Yetkisiz</span>
									{:else}
										<span class="status-tag tag-green">✅ Yetkili</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				<div class="panel-footer-hint">
					Whitelist yönetimi için <a href="/admin/whitelist">Admin Whitelist</a> sayfasını kullanın.
				</div>
			{/if}
		</div>
	</div>
	{/if}

	<!-- PANEL 2: Güvenlik Olayları -->
	{#if openSection === 'events'}
	<div class="panel">
		<div class="panel-header">
			<h3 class="panel-title">Güvenlik Olay Kaydı</h3>
			<div class="panel-filters">
				<select bind:value={eventFilter} class="filter-select">
					<option value="">Tümü</option>
					<option value="logon_success">Başarılı Giriş</option>
					<option value="logon_failed">Başarısız Giriş</option>
					<option value="logoff">Oturum Kapatma</option>
					<option value="user_created">Kullanıcı Oluşturma</option>
					<option value="group_member_added">Gruba Ekleme</option>
				</select>
			</div>
		</div>

		{#if summary && (summary.event_stats?.failed_logins || 0) >= 5}
			<div class="alert-banner alert-red" style="margin: 0; border-radius: 0;">
				🚨 Brute Force Alarmı: {summary.event_stats.failed_logins} başarısız giriş denemesi tespit edildi!
			</div>
		{/if}

		<div class="panel-body scrollable">
			{#if filteredEvents.length === 0}
				<div class="empty-state">Güvenlik olayı bulunamadı.</div>
			{:else}
				<table class="data-table">
					<thead>
						<tr>
							<th>Zaman</th>
							<th>Event ID</th>
							<th>Olay Tipi</th>
							<th>Kullanıcı</th>
							<th>IP Adresi</th>
							<th>Logon Type</th>
						</tr>
					</thead>
					<tbody>
						{#each filteredEvents as evt}
							<tr>
								<td class="mono">{formatTime(evt.event_time)}</td>
								<td class="mono">{evt.event_id}</td>
								<td>
									<span style="color: {eventTypeColor(evt.event_type)}; font-weight: 600; font-size: 11px;">
										{eventTypeLabel(evt.event_type)}
									</span>
								</td>
								<td class="mono">{evt.target_user || '-'}</td>
								<td class="mono">{evt.source_ip || '-'}</td>
								<td class="mono">{evt.logon_type ? logonTypeLabel(evt.logon_type) : '-'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
	{/if}

	<!-- PANEL 3: Dosya Bütünlüğü İzleme -->
	{#if openSection === 'fim'}
	<div class="panel">
		<div class="panel-header">
			<h3 class="panel-title">Dosya Bütünlüğü İzleme (FIM)</h3>
			<a href="/admin/fim-config" class="panel-link">Dosya Listesini Yönet →</a>
		</div>
		<div class="panel-body">
			{#if fimRecords.length === 0}
				<div class="empty-state">FIM verisi henüz alınmadı.</div>
			{:else}
				{#if fimChangedCount > 0}
					<div class="alert-banner alert-red">
						⚠ {fimChangedCount} dosyada bütünlük ihlali tespit edildi!
					</div>
				{/if}
				<table class="data-table">
					<thead>
						<tr>
							<th>Dosya Yolu</th>
							<th>SHA-256</th>
							<th>Boyut</th>
							<th>Durum</th>
							<th>Tarih</th>
						</tr>
					</thead>
					<tbody>
						{#each fimRecords as rec}
							<tr class:fim-changed={rec.status === 'changed'}>
								<td class="mono">{rec.file_path}</td>
								<td class="mono hash-cell" title={rec.sha256}>{shortHash(rec.sha256)}</td>
								<td class="mono">{rec.file_size ? (rec.file_size / 1024).toFixed(1) + ' KB' : '-'}</td>
								<td>
									<span class="status-tag" style="color: {fimStatusColor(rec.status)}">
										{fimStatusLabel(rec.status)}
									</span>
								</td>
								<td class="mono">{formatTime(rec.created_at)}</td>
							</tr>
							{#if rec.status === 'changed' && rec.previous_hash}
								<tr class="hash-compare-row">
									<td colspan="5">
										<div class="hash-compare">
											<span class="hash-old">Önceki: {rec.previous_hash}</span>
											<span class="hash-arrow">→</span>
											<span class="hash-new">Yeni: {rec.sha256}</span>
										</div>
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
	{/if}
</div>

<style>
	.health-bar {
		display: flex; align-items: center; gap: 16px;
		padding: 10px 16px; background: #14161e; border: 1px solid #1e2130;
		border-radius: 8px; margin-bottom: 16px;
	}
	.health-item {
		display: flex; align-items: center; gap: 6px; font-size: 12px; color: #64748b; font-weight: 500;
	}
	.health-dot { width: 8px; height: 8px; border-radius: 50%; background: #475569; }
	.health-item.healthy .health-dot { background: #22c55e; }
	.health-item.alert .health-dot { background: #ef4444; animation: blink 1.5s infinite; }
	@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
	.health-summary { margin-left: auto; font-size: 11px; color: #475569; }

	.section-tabs { display: flex; gap: 4px; margin-bottom: 16px; }
	.section-tab {
		padding: 8px 16px; background: #14161e; border: 1px solid #1e2130;
		border-radius: 6px; color: #64748b; font-size: 12px; font-weight: 500;
		cursor: pointer; transition: all 0.15s;
	}
	.section-tab:hover { border-color: #2d3348; color: #94a3b8; }
	.section-tab.active { background: #1e293b; border-color: #3b82f6; color: #60a5fa; }

	.panel {
		background: #14161e; border: 1px solid #1e2130; border-radius: 8px; overflow: hidden;
	}
	.panel-header {
		display: flex; justify-content: space-between; align-items: center;
		padding: 14px 18px; border-bottom: 1px solid #1e2130;
	}
	.panel-title { font-size: 14px; font-weight: 600; color: #94a3b8; margin: 0; }
	.panel-meta { font-size: 10px; color: #475569; font-family: monospace; }
	.panel-link { font-size: 11px; color: #60a5fa; text-decoration: none; font-weight: 500; }
	.panel-link:hover { text-decoration: underline; }
	.panel-body { padding: 0; }
	.scrollable { max-height: 500px; overflow-y: auto; }

	.panel-filters { display: flex; gap: 8px; }
	.filter-select {
		background: #0f1117; border: 1px solid #1e2130; color: #94a3b8;
		padding: 4px 10px; border-radius: 4px; font-size: 11px;
	}

	.alert-banner {
		padding: 10px 18px; font-size: 12px; font-weight: 600; margin: 12px 18px;
		border-radius: 6px;
	}
	.alert-red { background: rgba(239, 68, 68, 0.1); color: #ef4444; border: 1px solid rgba(239, 68, 68, 0.2); }

	.data-table { width: 100%; font-size: 12px; border-collapse: collapse; }
	.data-table thead { background: #0f1117; position: sticky; top: 0; }
	.data-table th {
		text-align: left; padding: 8px 14px; color: #64748b;
		font-weight: 600; text-transform: uppercase; font-size: 10px; letter-spacing: 0.3px;
		border-bottom: 1px solid #1e2130;
	}
	.data-table td { padding: 8px 14px; border-bottom: 1px solid #1a1d2a; color: #cbd5e1; }
	.data-table tbody tr:hover { background: #181b27; }
	.mono { font-family: monospace; font-size: 11px; }
	.hash-cell { font-size: 10px; color: #64748b; }

	.status-tag { font-size: 11px; font-weight: 600; }
	.tag-green { color: #22c55e; }
	.tag-red { color: #ef4444; }

	.unauthorized-row { background: rgba(239, 68, 68, 0.05); }
	.fim-changed { background: rgba(239, 68, 68, 0.05); }

	.hash-compare-row { background: rgba(239, 68, 68, 0.03); }
	.hash-compare {
		display: flex; align-items: center; gap: 8px;
		font-family: monospace; font-size: 10px; padding: 4px 0;
	}
	.hash-old { color: #ef4444; }
	.hash-arrow { color: #475569; }
	.hash-new { color: #f97316; }

	.empty-state { padding: 32px; text-align: center; color: #475569; font-size: 13px; }

	.panel-footer-hint {
		padding: 10px 18px; font-size: 11px; color: #475569;
		border-top: 1px solid #1e2130;
	}
	.panel-footer-hint a { color: #60a5fa; text-decoration: none; }
	.panel-footer-hint a:hover { text-decoration: underline; }

	.scrollable::-webkit-scrollbar { width: 5px; }
	.scrollable::-webkit-scrollbar-track { background: transparent; }
	.scrollable::-webkit-scrollbar-thumb { background: #2d3348; border-radius: 4px; }
</style>
