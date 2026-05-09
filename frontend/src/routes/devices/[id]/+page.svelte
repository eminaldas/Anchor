<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let id = $derived($page.params.id);

	let assetData = $state<any>(null);
	let openFindings = $state<any[]>([]);
	let closedFindings = $state<any[]>([]);
	let showClosed = $state(false);
	let closingId = $state('');

	// Severity → puan düşüşü eşlemesi
	const SCORE_DEDUCTION: Record<string, number> = {
		Critical: 30, High: 20, Medium: 10, Low: 5
	};

	// Uyum puanının neden düştüğünü açıkla
	let scoreBreakdown = $derived.by(() => {
		let score = 100;
		const lines: { label: string; deduction: number; severity: string }[] = [];
		for (const f of openFindings) {
			const d = SCORE_DEDUCTION[f.severity] ?? 0;
			if (d > 0) {
				score -= d;
				lines.push({ label: f.description, deduction: d, severity: f.severity });
			}
		}
		return { score: Math.max(0, score), lines };
	});

	async function fetchData() {
		try {
			const [aRes, openRes, closedRes] = await Promise.all([
				fetch(`${API}/assets/${id}`),
				fetch(`${API}/findings?asset_id=${id}&status=Open`),
				fetch(`${API}/findings?asset_id=${id}&status=Closed`)
			]);
			if (aRes.ok) assetData = await aRes.json();
			if (openRes.ok) {
				const d = await openRes.json();
				openFindings = d.data || [];
			}
			if (closedRes.ok) {
				const d = await closedRes.json();
				closedFindings = d.data || [];
			}
		} catch (err) { console.error(err); }
	}

	async function closeFinding(findingId: string) {
		closingId = findingId;
		try {
			await fetch(`${API}/findings/${findingId}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: 'Closed' })
			});
			await fetchData();
		} catch (err) { console.error(err); }
		finally { closingId = ''; }
	}

	async function reopenFinding(findingId: string) {
		closingId = findingId;
		try {
			await fetch(`${API}/findings/${findingId}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: 'Open' })
			});
			await fetchData();
		} catch (err) { console.error(err); }
		finally { closingId = ''; }
	}

	function formatUptime(seconds: number): string {
		if (!seconds) return 'N/A';
		const d = Math.floor(seconds / 86400);
		const h = Math.floor((seconds % 86400) / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		return `${d}g ${h}s ${m}dk`;
	}

	function checkTypeLabel(ct: string): string {
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
		fetchData();
		const interval = setInterval(fetchData, 10000);
		return () => clearInterval(interval);
	});
</script>

{#if assetData}
<div class="layout">

	<!-- Sol Kolon: Sistem + Güvenlik + Score -->
	<div class="left-col">
		<div class="panel">
			<h3 class="panel-title">Sistem Bilgisi</h3>
			<div class="info-list">
				<div class="info-row">
					<span class="info-label">CPU</span>
					<span class="info-value mono">{assetData.cpu_usage || 'N/A'}</span>
				</div>
				<div class="info-row">
					<span class="info-label">RAM</span>
					<span class="info-value mono">{assetData.ram_usage || 'N/A'}</span>
				</div>
				<div class="info-row">
					<span class="info-label">Uptime</span>
					<span class="info-value mono">{formatUptime(assetData.uptime)}</span>
				</div>
				<div class="info-row">
					<span class="info-label">Son Görülme</span>
					<span class="info-value mono">{new Date(assetData.last_seen).toLocaleString('tr-TR')}</span>
				</div>
			</div>
		</div>

		<div class="panel">
			<h3 class="panel-title">Güvenlik</h3>
			<div class="info-list">
				<div class="info-row">
					<span class="info-label">Windows Defender</span>
					{#if assetData.defender_enabled}
						<span class="badge badge-green">Aktif</span>
					{:else}
						<span class="badge badge-red">Kapalı</span>
					{/if}
				</div>
				<div class="info-row">
					<span class="info-label">Açık Bulgular</span>
					<span class="info-value mono">{openFindings.length}</span>
				</div>
				<div class="info-row">
					<span class="info-label">Kapatılan</span>
					<span class="info-value mono" style="color: #22c55e">{closedFindings.length}</span>
				</div>
			</div>
		</div>

		<!-- Uyum Puanı Breakdown -->
		<div class="panel score-panel">
			<h3 class="panel-title">Uyum Puanı Analizi</h3>
			<div class="score-display">
				<span class="score-number" class:good={assetData.compliance_score === 100} class:warn={assetData.compliance_score >= 70 && assetData.compliance_score < 100} class:bad={assetData.compliance_score < 70}>
					%{assetData.compliance_score ?? 100}
				</span>
				<div class="score-bar-wrap">
					<div class="score-bar" style="width: {assetData.compliance_score ?? 100}%; background: {(assetData.compliance_score ?? 100) === 100 ? '#22c55e' : (assetData.compliance_score ?? 100) >= 70 ? '#eab308' : '#ef4444'}"></div>
				</div>
			</div>
			{#if scoreBreakdown.lines.length === 0}
				<p class="score-ok">✅ Tüm kontroller geçildi. Puan tam.</p>
			{:else}
				<p class="score-label">Puanı düşüren bulgular:</p>
				<div class="score-items">
					<div class="score-item score-base">
						<span>Başlangıç puanı</span>
						<span class="mono">100</span>
					</div>
					{#each scoreBreakdown.lines as line}
						<div class="score-item">
							<span class="score-desc" title={line.label}>
								<span class="sev-dot sev-{line.severity.toLowerCase()}"></span>
								{checkTypeLabel(openFindings.find((f: any) => f.description === line.label)?.check_type || '')}
							</span>
							<span class="mono score-minus">−{line.deduction}</span>
						</div>
					{/each}
					<div class="score-item score-total">
						<span>Hesaplanan puan</span>
						<span class="mono" class:good={scoreBreakdown.score === 100} class:warn={scoreBreakdown.score >= 70 && scoreBreakdown.score < 100} class:bad={scoreBreakdown.score < 70}>
							{scoreBreakdown.score}
						</span>
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Sağ Kolon: Bulgular Yönetimi -->
	<div class="right-col">
		<div class="panel findings-panel">
			<div class="findings-header">
				<div>
					<h3 class="panel-title" style="margin-bottom: 0">Bulgular</h3>
					<span class="findings-sub">{openFindings.length} açık · {closedFindings.length} kapatıldı</span>
				</div>
				<div class="toggle-wrap">
					<button class="toggle-btn" class:active={!showClosed} onclick={() => showClosed = false}>
						Açık ({openFindings.length})
					</button>
					<button class="toggle-btn" class:active={showClosed} onclick={() => showClosed = true}>
						Kapatılanlar ({closedFindings.length})
					</button>
				</div>
			</div>

			{#if !showClosed}
				<!-- Açık Bulgular -->
				{#if openFindings.length === 0}
					<div class="empty-findings">
						<span class="empty-icon">✅</span>
						<p>Açık bulgu yok — sistem temiz!</p>
					</div>
				{:else}
					<div class="findings-list">
						{#each openFindings as f}
							<div class="finding-card finding-{f.severity.toLowerCase()}">
								<div class="finding-top">
									<div class="finding-meta">
										<span class="severity-badge sev-{f.severity.toLowerCase()}">{f.severity}</span>
										<span class="check-type">{checkTypeLabel(f.check_type)}</span>
										<span class="finding-time">{new Date(f.created_at).toLocaleString('tr-TR', { day:'2-digit', month:'2-digit', hour:'2-digit', minute:'2-digit' })}</span>
									</div>
									<button
										class="close-btn"
										onclick={() => closeFinding(f.id)}
										disabled={closingId === f.id}
										title="Bulguyu kapat"
									>
										{closingId === f.id ? '...' : 'Kapat'}
									</button>
								</div>
								<p class="finding-desc">{f.description}</p>
								<div class="finding-score-impact">
									<span>Puan etkisi:</span>
									<span class="mono score-minus">−{SCORE_DEDUCTION[f.severity] ?? 0} puan</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			{:else}
				<!-- Kapatılan Bulgular -->
				{#if closedFindings.length === 0}
					<div class="empty-findings">
						<p style="color: #475569">Henüz kapatılmış bulgu yok.</p>
					</div>
				{:else}
					<div class="findings-list">
						{#each closedFindings as f}
							<div class="finding-card finding-closed">
								<div class="finding-top">
									<div class="finding-meta">
										<span class="severity-badge sev-closed">{f.severity}</span>
										<span class="check-type" style="color: #475569">{checkTypeLabel(f.check_type)}</span>
										<span class="finding-time">{new Date(f.created_at).toLocaleString('tr-TR', { day:'2-digit', month:'2-digit', hour:'2-digit', minute:'2-digit' })}</span>
									</div>
									<button
										class="reopen-btn"
										onclick={() => reopenFinding(f.id)}
										disabled={closingId === f.id}
										title="Bulguyu yeniden aç"
									>
										{closingId === f.id ? '...' : 'Yeniden Aç'}
									</button>
								</div>
								<p class="finding-desc" style="color: #475569">{f.description}</p>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		</div>
	</div>

</div>
{:else}
<div class="loading-state">Yükleniyor...</div>
{/if}

<style>
	.layout {
		display: grid;
		grid-template-columns: 320px 1fr;
		gap: 20px;
		align-items: start;
	}
	.left-col { display: flex; flex-direction: column; gap: 16px; }
	.right-col { }

	.panel {
		background: #14161e; border: 1px solid #1e2130; border-radius: 8px;
		padding: 16px 18px;
	}
	.panel-title { font-size: 13px; font-weight: 600; color: #94a3b8; margin: 0 0 12px 0; text-transform: uppercase; letter-spacing: 0.4px; }

	.info-list { display: flex; flex-direction: column; }
	.info-row {
		display: flex; justify-content: space-between; align-items: center;
		padding: 8px 0; border-bottom: 1px solid #1a1d2a;
	}
	.info-row:last-child { border-bottom: none; }
	.info-label { font-size: 12px; color: #64748b; }
	.info-value { font-size: 12px; color: #e2e8f0; font-weight: 500; }
	.mono { font-family: monospace; font-size: 11px; }

	.badge { font-size: 10px; font-weight: 700; padding: 2px 8px; border-radius: 4px; }
	.badge-green { background: rgba(34,197,94,0.15); color: #22c55e; }
	.badge-red { background: rgba(239,68,68,0.15); color: #ef4444; }

	/* Score Panel */
	.score-panel { }
	.score-display { display: flex; align-items: center; gap: 14px; margin-bottom: 14px; }
	.score-number { font-size: 32px; font-weight: 700; font-family: monospace; min-width: 60px; }
	.score-number.good { color: #22c55e; }
	.score-number.warn { color: #eab308; }
	.score-number.bad { color: #ef4444; }
	.score-bar-wrap { flex: 1; height: 6px; background: #1e2130; border-radius: 4px; overflow: hidden; }
	.score-bar { height: 100%; border-radius: 4px; transition: width 0.4s ease; }

	.score-ok { font-size: 12px; color: #22c55e; margin: 0; }
	.score-label { font-size: 11px; color: #64748b; margin: 0 0 8px 0; }
	.score-items { display: flex; flex-direction: column; gap: 0; }
	.score-item {
		display: flex; justify-content: space-between; align-items: center;
		padding: 6px 8px; font-size: 11px; border-bottom: 1px solid #1a1d2a;
	}
	.score-item:last-child { border-bottom: none; }
	.score-base { color: #64748b; }
	.score-total { background: #0f1117; border-radius: 4px; font-weight: 700; margin-top: 4px; }
	.score-desc {
		display: flex; align-items: center; gap: 6px;
		color: #94a3b8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 180px;
	}
	.score-minus { color: #ef4444; font-weight: 700; }
	.sev-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
	.sev-dot.sev-critical { background: #ef4444; }
	.sev-dot.sev-high { background: #f97316; }
	.sev-dot.sev-medium { background: #eab308; }
	.sev-dot.sev-low { background: #3b82f6; }
	.good { color: #22c55e; }
	.warn { color: #eab308; }
	.bad { color: #ef4444; }

	/* Findings Panel */
	.findings-panel { padding: 0; overflow: hidden; }
	.findings-header {
		display: flex; justify-content: space-between; align-items: center;
		padding: 14px 18px; border-bottom: 1px solid #1e2130;
	}
	.findings-sub { font-size: 11px; color: #475569; margin-top: 2px; }
	.toggle-wrap { display: flex; background: #0f1117; border: 1px solid #1e2130; border-radius: 6px; overflow: hidden; }
	.toggle-btn {
		padding: 5px 14px; font-size: 11px; font-weight: 600; color: #64748b;
		border: none; cursor: pointer; background: transparent; transition: all 0.15s;
	}
	.toggle-btn.active { background: #1e293b; color: #60a5fa; }

	.findings-list { display: flex; flex-direction: column; gap: 0; max-height: 580px; overflow-y: auto; }
	.findings-list::-webkit-scrollbar { width: 5px; }
	.findings-list::-webkit-scrollbar-thumb { background: #2d3348; border-radius: 4px; }

	.finding-card {
		padding: 12px 18px; border-bottom: 1px solid #1a1d2a;
		border-left: 3px solid transparent;
	}
	.finding-card:hover { background: #181b27; }
	.finding-card.finding-critical { border-left-color: #ef4444; }
	.finding-card.finding-high { border-left-color: #f97316; }
	.finding-card.finding-medium { border-left-color: #eab308; }
	.finding-card.finding-low { border-left-color: #3b82f6; }
	.finding-card.finding-closed { border-left-color: #2d3348; opacity: 0.7; }

	.finding-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
	.finding-meta { display: flex; align-items: center; gap: 8px; }

	.severity-badge {
		font-size: 9px; font-weight: 700; padding: 2px 7px; border-radius: 3px;
		text-transform: uppercase; letter-spacing: 0.4px;
	}
	.sev-critical { background: rgba(239,68,68,0.15); color: #ef4444; }
	.sev-high { background: rgba(249,115,22,0.15); color: #f97316; }
	.sev-medium { background: rgba(234,179,8,0.15); color: #eab308; }
	.sev-low { background: rgba(59,130,246,0.15); color: #3b82f6; }
	.sev-closed { background: rgba(71,85,105,0.15); color: #475569; }

	.check-type { font-size: 12px; color: #94a3b8; font-weight: 500; }
	.finding-time { font-size: 10px; color: #475569; font-family: monospace; }

	.finding-desc { font-size: 12px; color: #cbd5e1; margin: 0 0 6px 0; line-height: 1.5; }
	.finding-score-impact { display: flex; align-items: center; gap: 6px; font-size: 10px; color: #475569; }

	.close-btn {
		padding: 4px 12px; background: rgba(239,68,68,0.1); color: #ef4444;
		border: 1px solid rgba(239,68,68,0.2); border-radius: 4px; font-size: 11px;
		font-weight: 600; cursor: pointer; transition: all 0.15s;
	}
	.close-btn:hover { background: rgba(239,68,68,0.2); }
	.close-btn:disabled { opacity: 0.5; cursor: wait; }

	.reopen-btn {
		padding: 4px 12px; background: rgba(59,130,246,0.1); color: #60a5fa;
		border: 1px solid rgba(59,130,246,0.2); border-radius: 4px; font-size: 11px;
		font-weight: 600; cursor: pointer; transition: all 0.15s;
	}
	.reopen-btn:hover { background: rgba(59,130,246,0.2); }
	.reopen-btn:disabled { opacity: 0.5; cursor: wait; }

	.empty-findings {
		padding: 48px 20px; text-align: center; color: #64748b; font-size: 13px;
	}
	.empty-icon { font-size: 28px; display: block; margin-bottom: 8px; }

	.loading-state { padding: 48px; text-align: center; color: #475569; }
</style>
