<script lang="ts">
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';

	let whitelist = $state<any[]>([]);
	let username = $state('');
	let reason = $state('');

	async function fetchWhitelist() {
		try {
			const res = await fetch(`${API}/admin/whitelist`);
			if (res.ok) whitelist = await res.json() || [];
		} catch (err) { console.error(err); }
	}

	async function addEntry(e: Event) {
		e.preventDefault();
		if (!username.trim()) return;

		try {
			const res = await fetch(`${API}/admin/whitelist`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username: username.trim(), reason, added_by: 'admin' })
			});
			if (res.ok) {
				username = '';
				reason = '';
				fetchWhitelist();
			} else {
				const data = await res.json();
				alert('Hata: ' + data.error);
			}
		} catch (err) { console.error(err); }
	}

	async function deleteEntry(id: number) {
		if (!confirm('Bu kaydı silmek istediğinize emin misiniz?')) return;
		try {
			const res = await fetch(`${API}/admin/whitelist/${id}`, { method: 'DELETE' });
			if (res.ok) fetchWhitelist();
		} catch (err) { console.error(err); }
	}

	onMount(fetchWhitelist);
</script>

<div class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Admin Whitelist Yönetimi</h1>
			<p class="page-subtitle">İzin verilen administrator hesaplarını yönetin. Bu listede olmayan adminler alarm üretir.</p>
		</div>
	</header>

	<div class="content-grid">
		<!-- Form -->
		<div class="panel form-panel">
			<h3 class="panel-title">Yeni Kayıt Ekle</h3>
			<form onsubmit={addEntry} class="form">
				<div class="form-group">
					<label class="form-label">Kullanıcı Adı</label>
					<input type="text" bind:value={username} placeholder="Ör: Administrator" class="form-input" required />
				</div>
				<div class="form-group">
					<label class="form-label">Sebep / Açıklama</label>
					<input type="text" bind:value={reason} placeholder="Ör: Windows varsayılan admin" class="form-input" />
				</div>
				<button type="submit" class="form-btn">Ekle</button>
			</form>
		</div>

		<!-- Liste -->
		<div class="panel">
			<div class="panel-header">
				<h3 class="panel-title">Mevcut Whitelist ({whitelist.length})</h3>
			</div>
			<div class="panel-body">
				{#if whitelist.length === 0}
					<div class="empty-state">Whitelist boş.</div>
				{:else}
					<table class="data-table">
						<thead>
							<tr>
								<th>Kullanıcı Adı</th>
								<th>Sebep</th>
								<th>Ekleyen</th>
								<th>Tarih</th>
								<th style="text-align: right">İşlem</th>
							</tr>
						</thead>
						<tbody>
							{#each whitelist as entry}
								<tr>
									<td class="mono">{entry.username}</td>
									<td>{entry.reason || '-'}</td>
									<td>{entry.added_by || '-'}</td>
									<td class="mono">{new Date(entry.created_at).toLocaleDateString('tr-TR')}</td>
									<td style="text-align: right">
										<button onclick={() => deleteEntry(entry.id)} class="delete-btn">Sil</button>
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

	.content-grid { display: grid; grid-template-columns: 300px 1fr; gap: 20px; align-items: start; }

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
		transition: background 0.15s;
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
