<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let id = $derived($page.params.id);

	let commands = $state<any[]>([]);
	let selectedResult = $state<any>(null);
	let resultModalOpen = $state(false);
	let sendingCmd = $state('');

	async function fetchCommands() {
		try {
			const res = await fetch(`${API}/commands?asset_id=${id}`);
			if (res.ok) commands = await res.json() || [];
		} catch (err) { console.error(err); }
	}

	async function sendCommand(type: string) {
		sendingCmd = type;
		try {
			await fetch(`${API}/commands`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ asset_id: id, type, params: '' })
			});
			await fetchCommands();
		} catch (err) { console.error(err); }
		finally { sendingCmd = ''; }
	}

	function viewResult(cmd: any) { selectedResult = cmd; resultModalOpen = true; }

	function statusIcon(s: string) { return { pending: '🕐', sent: '⏳', completed: '✅', failed: '❌' }[s] || '❓'; }
	function statusColor(s: string) { return { pending: '#eab308', sent: '#3b82f6', completed: '#22c55e', failed: '#ef4444' }[s] || '#6b7280'; }

	function formatDate(d: string) {
		return new Date(d).toLocaleString('tr-TR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	onMount(() => {
		fetchCommands();
		const interval = setInterval(fetchCommands, 5000);
		return () => clearInterval(interval);
	});
</script>

<div class="cmd-layout">
	<!-- Hızlı Aksiyonlar -->
	<div class="actions-panel">
		<h3 class="panel-title">Hızlı Aksiyonlar</h3>
		<div class="action-grid">
			<button onclick={() => sendCommand('screenshot')} disabled={sendingCmd === 'screenshot'} class="action-btn">
				<span class="action-icon">📸</span>
				<div>
					<div class="action-label">Ekran Görüntüsü</div>
					<div class="action-desc">Anlık ekran yakalama</div>
				</div>
				{#if sendingCmd === 'screenshot'}<span class="spin">⏳</span>{/if}
			</button>
			<button onclick={() => sendCommand('list_processes')} disabled={sendingCmd === 'list_processes'} class="action-btn">
				<span class="action-icon">📋</span>
				<div>
					<div class="action-label">Süreç Listesi</div>
					<div class="action-desc">Çalışan process'ler</div>
				</div>
				{#if sendingCmd === 'list_processes'}<span class="spin">⏳</span>{/if}
			</button>
			<button onclick={() => sendCommand('sysinfo')} disabled={sendingCmd === 'sysinfo'} class="action-btn">
				<span class="action-icon">🖥️</span>
				<div>
					<div class="action-label">Sistem Bilgisi</div>
					<div class="action-desc">OS, Disk, Ağ detayları</div>
				</div>
				{#if sendingCmd === 'sysinfo'}<span class="spin">⏳</span>{/if}
			</button>
		</div>
		<p class="action-hint">Komutlar ajanın bir sonraki heartbeat'inde iletilir (~10sn)</p>
	</div>

	<!-- Komut Geçmişi -->
	<div class="panel">
		<div class="panel-header">
			<h3 class="panel-title">Komut Geçmişi ({commands.length})</h3>
		</div>
		<div class="panel-body scrollable">
			{#if commands.length === 0}
				<div class="empty-state">Henüz komut gönderilmedi.</div>
			{:else}
				<table class="data-table">
					<thead>
						<tr>
							<th>Tip</th>
							<th>Durum</th>
							<th>Tarih</th>
							<th style="text-align: right">İşlem</th>
						</tr>
					</thead>
					<tbody>
						{#each commands as cmd}
							<tr>
								<td class="mono">{cmd.type}</td>
								<td>
									<span style="color: {statusColor(cmd.status)}; font-weight: 600; font-size: 11px;">
										{statusIcon(cmd.status)} {cmd.status}
									</span>
								</td>
								<td class="mono">{formatDate(cmd.created_at)}</td>
								<td style="text-align: right">
									{#if cmd.status === 'completed'}
										<button onclick={() => viewResult(cmd)} class="result-btn btn-blue">Sonuç</button>
									{:else if cmd.status === 'failed'}
										<button onclick={() => viewResult(cmd)} class="result-btn btn-red">Hata</button>
									{:else}
										<span class="waiting-text">Bekleniyor...</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
</div>

<!-- SONUÇ MODAL -->
{#if resultModalOpen && selectedResult}
<div class="modal-overlay" onclick={() => resultModalOpen = false}>
	<div class="modal-content" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h3>{statusIcon(selectedResult.status)} Komut Sonucu: <span class="mono">{selectedResult.type}</span></h3>
			<button onclick={() => resultModalOpen = false} class="modal-close">✕</button>
		</div>
		<div class="modal-body scrollable">
			{#if selectedResult.type === 'screenshot' && selectedResult.status === 'completed'}
				<img src="data:image/png;base64,{selectedResult.result}" alt="Ekran Görüntüsü" class="screenshot-img" />
			{:else if selectedResult.type === 'list_processes' && selectedResult.status === 'completed'}
				{@const procs = JSON.parse(selectedResult.result)}
				<table class="data-table">
					<thead><tr><th>PID</th><th>İsim</th><th style="text-align:right">CPU%</th><th style="text-align:right">RAM%</th></tr></thead>
					<tbody>
						{#each procs.sort((a: any, b: any) => b.cpu - a.cpu).slice(0, 100) as p}
							<tr><td class="mono">{p.pid}</td><td>{p.name}</td><td class="mono" style="text-align:right">{p.cpu.toFixed(1)}</td><td class="mono" style="text-align:right">{p.ram.toFixed(1)}</td></tr>
						{/each}
					</tbody>
				</table>
			{:else if selectedResult.type === 'sysinfo' && selectedResult.status === 'completed'}
				{@const info = JSON.parse(selectedResult.result)}
				<div class="sysinfo-grid">
					{#each Object.entries({OS: info.os, Platform: info.platform, Hostname: info.hostname, Kernel: info.kernel_version, CPU: info.cpu_model, 'CPU Cores': info.cpu_cores, 'Toplam RAM': info.total_ram}) as [k, v]}
						<div class="sysinfo-item"><span class="sysinfo-label">{k}</span><span class="mono">{v}</span></div>
					{/each}
				</div>
				{#if info.disks?.length}
					<h4 class="sub-title">Diskler</h4>
					{#each info.disks as d}
						<div class="sysinfo-item"><span class="mono">{d.device} ({d.mount})</span><span class="mono" style="color: #64748b">{d.used} / {d.total}</span></div>
					{/each}
				{/if}
			{:else}
				<pre class="result-pre">{selectedResult.result}</pre>
			{/if}
		</div>
	</div>
</div>
{/if}

<style>
	.cmd-layout { display: flex; flex-direction: column; gap: 20px; }

	.actions-panel {
		background: #14161e; border: 1px solid #1e2130; border-radius: 8px; padding: 18px 20px;
	}
	.panel-title { font-size: 14px; font-weight: 600; color: #94a3b8; margin: 0 0 14px 0; }
	.action-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
	.action-btn {
		display: flex; align-items: center; gap: 10px;
		padding: 12px 14px; background: #0f1117; border: 1px solid #1e2130;
		border-radius: 6px; color: #e2e8f0; cursor: pointer;
		transition: border-color 0.15s; text-align: left;
	}
	.action-btn:hover { border-color: #3b82f6; }
	.action-btn:disabled { opacity: 0.5; cursor: wait; }
	.action-icon { font-size: 20px; }
	.action-label { font-size: 13px; font-weight: 600; }
	.action-desc { font-size: 10px; color: #64748b; }
	.spin { margin-left: auto; animation: spin 1s linear infinite; }
	@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
	.action-hint { font-size: 10px; color: #475569; text-align: center; margin-top: 10px; }

	.panel { background: #14161e; border: 1px solid #1e2130; border-radius: 8px; overflow: hidden; }
	.panel-header { display: flex; justify-content: space-between; align-items: center; padding: 14px 18px; border-bottom: 1px solid #1e2130; }
	.panel-body { padding: 0; }
	.scrollable { max-height: 400px; overflow-y: auto; }

	.data-table { width: 100%; font-size: 12px; border-collapse: collapse; }
	.data-table thead { background: #0f1117; position: sticky; top: 0; }
	.data-table th { text-align: left; padding: 8px 14px; color: #64748b; font-weight: 600; text-transform: uppercase; font-size: 10px; border-bottom: 1px solid #1e2130; }
	.data-table td { padding: 8px 14px; border-bottom: 1px solid #1a1d2a; color: #cbd5e1; }
	.data-table tbody tr:hover { background: #181b27; }
	.mono { font-family: monospace; font-size: 11px; }

	.result-btn { padding: 4px 12px; border-radius: 4px; font-size: 11px; font-weight: 600; cursor: pointer; border: none; }
	.btn-blue { background: #1e3a5f; color: #60a5fa; }
	.btn-blue:hover { background: #1e4a7f; }
	.btn-red { background: #3f1515; color: #ef4444; }
	.btn-red:hover { background: #5f1515; }
	.waiting-text { font-size: 10px; color: #475569; }

	.modal-overlay {
		position: fixed; inset: 0; background: rgba(0,0,0,0.7);
		display: flex; align-items: center; justify-content: center; z-index: 50; padding: 20px;
	}
	.modal-content {
		background: #14161e; border: 1px solid #2d3348; border-radius: 10px;
		max-width: 900px; width: 100%; max-height: 80vh; display: flex; flex-direction: column;
	}
	.modal-header {
		display: flex; justify-content: space-between; align-items: center;
		padding: 16px 20px; border-bottom: 1px solid #1e2130; font-size: 15px; font-weight: 600; color: #e2e8f0;
	}
	.modal-close { background: none; border: none; color: #64748b; font-size: 18px; cursor: pointer; }
	.modal-close:hover { color: #e2e8f0; }
	.modal-body { padding: 20px; overflow-y: auto; flex: 1; }

	.screenshot-img { width: 100%; border-radius: 6px; border: 1px solid #1e2130; }
	.result-pre { background: #0f1117; padding: 16px; border-radius: 6px; font-size: 11px; color: #cbd5e1; font-family: monospace; white-space: pre-wrap; word-break: break-all; }

	.sysinfo-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
	.sysinfo-item { background: #0f1117; padding: 10px 14px; border-radius: 6px; border: 1px solid #1e2130; display: flex; justify-content: space-between; align-items: center; }
	.sysinfo-label { font-size: 11px; color: #64748b; }
	.sub-title { font-size: 13px; font-weight: 600; color: #94a3b8; margin: 16px 0 8px; }

	.empty-state { padding: 32px; text-align: center; color: #475569; font-size: 13px; }

	.scrollable::-webkit-scrollbar { width: 5px; }
	.scrollable::-webkit-scrollbar-track { background: transparent; }
	.scrollable::-webkit-scrollbar-thumb { background: #2d3348; border-radius: 4px; }
</style>
