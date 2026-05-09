<script lang="ts">
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';

	let configs = $state<any[]>([]);
	let filePath = $state('');
	let label = $state('');
	let priority = $state('High');

	async function fetchConfigs() {
		try {
			const res = await fetch(`${API}/admin/fim-config`);
			if (res.ok) configs = await res.json() || [];
		} catch (err) { console.error(err); }
	}

	async function addConfig(e: Event) {
		e.preventDefault();
		if (!filePath.trim()) return;

		try {
			const res = await fetch(`${API}/admin/fim-config`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ file_path: filePath.trim(), label, priority })
			});
			if (res.ok) {
				filePath = '';
				label = '';
				priority = 'High';
				fetchConfigs();
			} else {
				const data = await res.json();
				alert('Hata: ' + data.error);
			}
		} catch (err) { console.error(err); }
	}

	async function deleteConfig(id: number) {
		if (!confirm('Bu yapılandırmayı silmek istediğinize emin misiniz?')) return;
		try {
			const res = await fetch(`${API}/admin/fim-config/${id}`, { method: 'DELETE' });
			if (res.ok) fetchConfigs();
		} catch (err) { console.error(err); }
	}

	function priorityColor(p: string) {
		return { Critical: '#ef4444', High: '#f97316', Medium: '#eab308' }[p] || '#6b7280';
	}

	onMount(fetchConfigs);
</script>

<div class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">FIM Yapılandırma</h1>
			<p class="page-subtitle">Dosya Bütünlüğü İzleme (FIM) için takip edilecek kritik sistem dosyalarını yönetin.</p>
		</div>
	</header>

	<div class="content-grid">
		<!-- Form -->
		<div class="panel form-panel">
			<h3 class="panel-title">Yeni Dosya Ekle</h3>
			<form onsubmit={addConfig} class="form">
				<div class="form-group">
					<label class="form-label">Dosya Yolu</label>
					<input type="text" bind:value={filePath} placeholder="Ör: C:\Windows\System32\drivers\etc\hosts" class="form-input" required />
				</div>
				<div class="form-group">
					<label class="form-label">Etiket</label>
					<input type="text" bind:value={label} placeholder="Ör: DNS Yapılandırması" class="form-input" />
				</div>
				<div class="form-group">
					<label class="form-label">Öncelik</label>
					<select bind:value={priority} class="form-input">
						<option value="Critical">Kritik</option>
						<option value="High">Yüksek</option>
						<option value="Medium">Orta</option>
					</select>
				</div>
				<button type="submit" class="form-btn">Ekle</button>
			</form>
		</div>

		<!-- Liste -->
		<div class="panel">
			<div class="panel-header">
				<h3 class="panel-title">İzlenen Dosyalar ({configs.length})</h3>
			</div>
			<div class="panel-body">
				{#if configs.length === 0}
					<div class="empty-state">Yapılandırma bulunmuyor.</div>
				{:else}
					<table class="data-table">
						<thead>
							<tr>
								<th>Dosya Yolu</th>
								<th>Etiket</th>
								<th>Öncelik</th>
								<th>Tarih</th>
								<th style="text-align: right">İşlem</th>
							</tr>
						</thead>
						<tbody>
							{#each configs as cfg}
								<tr>
									<td class="mono">{cfg.file_path}</td>
									<td>{cfg.label || '-'}</td>
									<td>
										<span style="color: {priorityColor(cfg.priority)}; font-weight: 600; font-size: 11px;">
											{cfg.priority}
										</span>
									</td>
									<td class="mono">{new Date(cfg.created_at).toLocaleDateString('tr-TR')}</td>
									<td style="text-align: right">
										<button onclick={() => deleteConfig(cfg.id)} class="delete-btn">Sil</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.page { padding: 28px 32px; color: #e2e8f0; }
	.page-header { margin-bottom: 24px; padding-bottom: 16px; border-bottom: 1px solid #1e2130; }
	.page-title { font-size: 22px; font-weight: 700; color: #f1f5f9; margin: 0; }
	.page-subtitle { font-size: 13px; color: #64748b; margin-top: 4px; }

	.content-grid { display: grid; grid-template-columns: 320px 1fr; gap: 20px; align-items: start; }

	.panel { background: #14161e; border: 1px solid #1e2130; border-radius: 8px; overflow: hidden; }
	.form-panel { padding: 18px 20px; }
	.panel-header { padding: 14px 18px; border-bottom: 1px solid #1e2130; }
	.panel-title { font-size: 14px; font-weight: 600; color: #94a3b8; margin: 0; }
	.panel-body { padding: 0; }

	.form { display: flex; flex-direction: column; gap: 12px; margin-top: 14px; }
	.form-group { display: flex; flex-direction: column; gap: 4px; }
	.form-label { font-size: 11px; color: #64748b; font-weight: 600; text-transform: uppercase; }
	.form-input {
		background: #0f1117; border: 1px solid #1e2130; color: #e2e8f0;
		padding: 8px 12px; border-radius: 6px; font-size: 13px;
	}
	.form-input:focus { outline: none; border-color: #3b82f6; }
	.form-btn {
		padding: 8px 16px; background: #1e3a5f; color: #60a5fa; border: 1px solid #2d4a7f;
		border-radius: 6px; font-size: 13px; font-weight: 600; cursor: pointer;
	}
	.form-btn:hover { background: #1e4a7f; }

	.data-table { width: 100%; font-size: 12px; border-collapse: collapse; }
	.data-table th {
		text-align: left; padding: 8px 14px; color: #64748b;
		font-weight: 600; text-transform: uppercase; font-size: 10px;
		border-bottom: 1px solid #1e2130; background: #0f1117;
	}
	.data-table td { padding: 8px 14px; border-bottom: 1px solid #1a1d2a; color: #cbd5e1; }
	.data-table tbody tr:hover { background: #181b27; }
	.mono { font-family: monospace; font-size: 11px; }

	.delete-btn {
		background: rgba(239, 68, 68, 0.1); color: #ef4444; border: none;
		padding: 4px 12px; border-radius: 4px; font-size: 11px; font-weight: 600;
		cursor: pointer;
	}
	.delete-btn:hover { background: rgba(239, 68, 68, 0.2); }

	.empty-state { padding: 32px; text-align: center; color: #475569; font-size: 13px; }
</style>
