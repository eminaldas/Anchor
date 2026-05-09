<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	const API = 'http://localhost:8081/api/v1';
	let id = $derived($page.params.id);

	let assetData = $state<any>(null);
	let policies = $state<any[]>([]);

	let softwareList = $derived.by(() => {
		if (!assetData || !assetData.software_list) return [];
		if (Array.isArray(assetData.software_list)) return assetData.software_list;
		try { return JSON.parse(assetData.software_list); } catch { return []; }
	});

	let policyMap = $derived.by(() => {
		const m: Record<string, any> = {};
		for (const p of policies) m[p.name] = p;
		return m;
	});

	let bannedCount = $derived(softwareList.filter((s: string) => policyMap[s]?.status === 'Banned').length);

	async function fetchData() {
		try {
			const [aRes, pRes] = await Promise.all([
				fetch(`${API}/assets/${id}`),
				fetch(`${API}/policies`)
			]);
			if (aRes.ok) assetData = await aRes.json();
			if (pRes.ok) policies = await pRes.json() || [];
		} catch (err) { console.error(err); }
	}

	onMount(() => { fetchData(); });
</script>

<div>
	<div class="toolbar">
		<span class="toolbar-info">{softwareList.length} yazılım tespit edildi</span>
		{#if bannedCount > 0}
			<span class="toolbar-alert">{bannedCount} yasaklı yazılım!</span>
		{/if}
		<a href="/admin/policies" class="toolbar-link">Politikaları Yönet →</a>
	</div>

	{#if softwareList.length === 0}
		<div class="empty-state">Yazılım envanteri henüz alınmadı.</div>
	{:else}
		<div class="software-list">
			{#each softwareList as software}
				{@const policy = policyMap[software]}
				<div class="software-item" class:banned={policy?.status === 'Banned'} class:approved={policy?.status === 'Approved'}>
					<span class="sw-name">{software}</span>
					<div class="sw-tags">
						{#if policy?.status === 'Banned'}
							<span class="sw-reason" title={policy.reason}>{policy.reason}</span>
							<span class="sw-badge badge-banned">YASAKLI</span>
						{:else if policy?.status === 'Approved'}
							<span class="sw-badge badge-approved">Güvenli</span>
						{:else if policy}
							<span class="sw-badge badge-ignored">Gözardı</span>
						{:else}
							<span class="sw-badge badge-unknown">İncelenmedi</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.toolbar {
		display: flex; align-items: center; gap: 12px;
		padding: 10px 16px; background: #14161e;
		border: 1px solid #1e2130; border-radius: 8px;
		margin-bottom: 16px;
	}
	.toolbar-info { font-size: 13px; color: #94a3b8; font-weight: 500; }
	.toolbar-alert {
		font-size: 11px; font-weight: 700; color: #ef4444;
		background: rgba(239, 68, 68, 0.1); padding: 3px 10px; border-radius: 4px;
	}
	.toolbar-link {
		margin-left: auto; font-size: 12px; color: #60a5fa; text-decoration: none; font-weight: 500;
	}
	.toolbar-link:hover { text-decoration: underline; }

	.software-list { display: flex; flex-direction: column; gap: 4px; }
	.software-item {
		display: flex; justify-content: space-between; align-items: center;
		padding: 10px 16px; background: #14161e; border: 1px solid #1e2130;
		border-radius: 6px; transition: border-color 0.15s;
	}
	.software-item:hover { border-color: #2d3348; }
	.software-item.banned { background: rgba(239, 68, 68, 0.05); border-color: rgba(239, 68, 68, 0.2); }
	.software-item.approved { border-color: rgba(34, 197, 94, 0.15); }

	.sw-name { font-size: 13px; color: #cbd5e1; }
	.sw-tags { display: flex; align-items: center; gap: 8px; }
	.sw-reason { font-size: 10px; color: #64748b; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	.sw-badge {
		font-size: 9px; font-weight: 700; padding: 2px 8px; border-radius: 3px;
		text-transform: uppercase; letter-spacing: 0.5px;
	}
	.badge-banned { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
	.badge-approved { background: rgba(34, 197, 94, 0.15); color: #22c55e; }
	.badge-ignored { background: rgba(107, 114, 128, 0.15); color: #6b7280; }
	.badge-unknown { background: rgba(71, 85, 105, 0.15); color: #475569; }

	.empty-state { padding: 48px; text-align: center; color: #475569; background: #14161e; border: 1px solid #1e2130; border-radius: 8px; }
</style>
