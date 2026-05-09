<script lang="ts">
    import { onMount } from 'svelte';

    let activeTab = $state('catalog'); // 'catalog' or 'policies'
    
    let catalog = $state<any[]>([]);
    let policies = $state<any[]>([]);
    
    // Bulk action states
    let selectedApps = $state<Set<string>>(new Set());
    let bulkRisk = $state("High");
    let bulkReason = $state("");

    // Tekil ekleme states (Aktif kurallar sekmesi için opsiyonel)
    let singleSoftware = $state("");
    let singleRisk = $state("High");
    let singleReason = $state("");

    // Her uygulamanın o anki durumunu hızlıca bulmak için map
    let policyMap = $derived.by(() => {
        const m: Record<string, any> = {};
        for (const p of policies) {
            m[p.name] = p;
        }
        return m;
    });

    function isNew(dateString: string) {
        const diff = Date.now() - new Date(dateString).getTime();
        return diff < 24 * 60 * 60 * 1000; // 24 hours
    }

    async function fetchData() {
        try {
            const [catRes, polRes] = await Promise.all([
                fetch('http://localhost:8081/api/v1/inventory/catalog'),
                fetch('http://localhost:8081/api/v1/policies')
            ]);
            catalog = await catRes.json() || [];
            policies = await polRes.json() || [];
        } catch (err) {
            console.error("Veri çekilemedi:", err);
        }
    }

    function toggleSelection(appName: string) {
        if (selectedApps.has(appName)) {
            selectedApps.delete(appName);
        } else {
            selectedApps.add(appName);
        }
        // state trigger
        selectedApps = new Set(selectedApps);
    }

    function toggleAll() {
        if (selectedApps.size === catalog.length) {
            selectedApps.clear();
        } else {
            for (const c of catalog) selectedApps.add(c.name);
        }
        selectedApps = new Set(selectedApps);
    }

    async function applyBulkAction(action: 'ban' | 'approve') {
        if (selectedApps.size === 0) return alert("Lütfen en az bir uygulama seçin.");
        if (action === 'ban' && !bulkReason) return alert("Lütfen yasaklama için bir sebep girin.");

        try {
            const res = await fetch('http://localhost:8081/api/v1/policies/bulk', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    apps: Array.from(selectedApps),
                    action: action,
                    risk_level: bulkRisk,
                    reason: bulkReason || 'Otomatik İzin'
                })
            });

            if (res.ok) {
                selectedApps.clear();
                bulkReason = "";
                await fetchData();
            } else {
                const data = await res.json();
                alert("Hata: " + data.error);
            }
        } catch (err) {
            console.error("Toplu işlem başarısız:", err);
        }
    }

    async function addSinglePolicy(e: Event) {
        e.preventDefault();
        if (!singleSoftware) return alert("Lütfen bir yazılım adı girin.");

        try {
            const res = await fetch('http://localhost:8081/api/v1/policies/bulk', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    apps: [singleSoftware],
                    action: 'ban',
                    risk_level: singleRisk,
                    reason: singleReason
                })
            });

            if (res.ok) {
                singleSoftware = "";
                singleReason = "";
                await fetchData();
            } else {
                const data = await res.json();
                alert("Hata: " + data.error);
            }
        } catch (err) {
            console.error("Politika eklenemedi:", err);
        }
    }

    async function deletePolicy(id: number) {
        if (!confirm("Bu politikayı silmek istediğinize emin misiniz?")) return;
        try {
            const res = await fetch(`http://localhost:8081/api/v1/policies/${id}`, { method: 'DELETE' });
            if (res.ok) fetchData();
        } catch (err) {
            console.error("Politika silinemedi:", err);
        }
    }

    onMount(fetchData);
</script>

<div class="layout-container">
    <header class="page-header">
        <div>
            <h1 class="page-title">Yazılım & Politika Yönetimi</h1>
            <p class="page-desc">Tüm cihazlardaki yazılım envanterini (Katalog) görün ve toplu kurallar uygulayın.</p>
        </div>
    </header>

    <!-- Tabs -->
    <div class="tabs">
        <button class="tab" class:active={activeTab === 'catalog'} onclick={() => activeTab = 'catalog'}>
            Yazılım Kataloğu ({catalog.length})
        </button>
        <button class="tab" class:active={activeTab === 'policies'} onclick={() => activeTab = 'policies'}>
            Aktif Kurallar ({policies.length})
        </button>
    </div>

    {#if activeTab === 'catalog'}
        <!-- Katalog Sekmesi -->
        <div class="panel">
            <div class="bulk-actions">
                <div class="bulk-form">
                    <select bind:value={bulkRisk} class="input-field" style="width: 150px;">
                        <option value="Critical">Kritik Risk</option>
                        <option value="High">Yüksek Risk</option>
                        <option value="Medium">Orta Risk</option>
                        <option value="Low">Düşük Risk</option>
                    </select>
                    <input 
                        type="text" 
                        bind:value={bulkReason} 
                        placeholder="Yasaklama sebebi (Örn: Shadow IT, Lisans İhlali...)" 
                        class="input-field" 
                        list="reasons-list"
                        style="flex: 1;"
                    />
                    <datalist id="reasons-list">
                        <option value="Kurumsal Onay Dışı (Shadow IT)"></option>
                        <option value="Veri Sızıntısı (Data Leakage) Riski"></option>
                        <option value="Zararlı Yazılım Şüphesi"></option>
                        <option value="Lisans İhlali"></option>
                    </datalist>
                </div>
                <div class="bulk-buttons">
                    <span class="selected-count">{selectedApps.size} seçildi</span>
                    <button class="btn btn-danger" onclick={() => applyBulkAction('ban')} disabled={selectedApps.size === 0}>
                        🚫 Seçilenleri Yasakla
                    </button>
                    <button class="btn btn-success" onclick={() => applyBulkAction('approve')} disabled={selectedApps.size === 0}>
                        ✅ Seçilenlere İzin Ver
                    </button>
                </div>
            </div>

            <div class="table-container">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th style="width: 40px; text-align: center;">
                                <input type="checkbox" checked={selectedApps.size === catalog.length && catalog.length > 0} onclick={toggleAll} />
                            </th>
                            <th>Yazılım Adı</th>
                            <th>Mevcut Durum</th>
                            <th>İlk Görülme</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#if catalog.length === 0}
                            <tr>
                                <td colspan="4" class="empty-state">Katalogda henüz yazılım yok.</td>
                            </tr>
                        {/if}
                        {#each catalog as item}
                            <tr class:selected={selectedApps.has(item.name)}>
                                <td style="text-align: center;">
                                    <input type="checkbox" checked={selectedApps.has(item.name)} onchange={() => toggleSelection(item.name)} />
                                </td>
                                <td class="mono">
                                    {item.name}
                                    {#if isNew(item.first_seen_at)}
                                        <span class="badge" style="background: rgba(34,197,94,0.15); color: #22c55e; margin-left: 8px;">YENİ</span>
                                    {/if}
                                </td>
                                <td>
                                    {#if policyMap[item.name]}
                                        <span class="badge badge-red">Yasaklı ({policyMap[item.name].risk_level})</span>
                                    {:else}
                                        <span class="badge badge-gray">İzinli / Kural Yok</span>
                                    {/if}
                                </td>
                                <td class="mono" style="color: #64748b;">
                                    {new Date(item.first_seen_at).toLocaleDateString('tr-TR')}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}

    {#if activeTab === 'policies'}
        <!-- Aktif Kurallar Sekmesi -->
        <div class="panel-grid">
            <!-- Manuel Kural Ekleme -->
            <div class="panel h-fit">
                <h2 class="panel-title border-bottom">Manuel Kural Ekle</h2>
                <form onsubmit={addSinglePolicy} class="form-vertical">
                    <label>Yazılım Adı (Tam Eşleşme)</label>
                    <input type="text" bind:value={singleSoftware} class="input-field" placeholder="Örn: Steam" required />

                    <label>Risk Seviyesi</label>
                    <select bind:value={singleRisk} class="input-field">
                        <option value="Critical">Kritik</option>
                        <option value="High">Yüksek</option>
                        <option value="Medium">Orta</option>
                        <option value="Low">Düşük</option>
                    </select>

                    <label>Sebep</label>
                    <input type="text" bind:value={singleReason} class="input-field" required />

                    <button type="submit" class="btn btn-primary" style="margin-top: 10px;">Kuralı Kaydet</button>
                </form>
            </div>

            <!-- Kurallar Listesi -->
            <div class="panel">
                <h2 class="panel-title border-bottom">Yasaklı Yazılımlar</h2>
                <div class="table-container">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>Yazılım</th>
                                <th>Risk</th>
                                <th>Sebep</th>
                                <th style="text-align: right;">İşlem</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#if policies.length === 0}
                                <tr>
                                    <td colspan="4" class="empty-state">Hiç yasaklı kural bulunmuyor.</td>
                                </tr>
                            {/if}
                            {#each policies as policy}
                                <tr>
                                    <td class="mono">{policy.name}</td>
                                    <td>
                                        <span class="badge {policy.risk_level === 'Critical' ? 'badge-red' : 'badge-orange'}">
                                            {policy.risk_level}
                                        </span>
                                    </td>
                                    <td>{policy.reason}</td>
                                    <td style="text-align: right;">
                                        <button onclick={() => deletePolicy(policy.id)} class="btn-icon text-red" title="Kuralı Sil">🗑️ Sil</button>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    {/if}

</div>

<style>
    .layout-container {
        padding: 24px;
        color: #e2e8f0;
    }
    .page-header {
        margin-bottom: 24px;
        padding-bottom: 16px;
        border-bottom: 1px solid #1e2130;
    }
    .page-title { font-size: 24px; font-weight: 700; color: #f87171; margin: 0 0 8px 0; }
    .page-desc { color: #94a3b8; font-size: 14px; margin: 0; }

    /* Tabs */
    .tabs { display: flex; gap: 4px; margin-bottom: 20px; border-bottom: 1px solid #1e2130; }
    .tab {
        padding: 10px 20px; background: transparent; border: none;
        color: #64748b; font-weight: 600; font-size: 14px; cursor: pointer;
        border-bottom: 2px solid transparent; transition: all 0.2s;
    }
    .tab:hover { color: #e2e8f0; }
    .tab.active { color: #60a5fa; border-bottom-color: #60a5fa; }

    /* Panels */
    .panel {
        background: #14161e; border: 1px solid #1e2130; border-radius: 8px;
        padding: 20px; margin-bottom: 20px;
    }
    .panel-grid { display: grid; grid-template-columns: 300px 1fr; gap: 20px; align-items: start; }
    .h-fit { height: fit-content; }
    .panel-title { font-size: 16px; font-weight: 600; color: #e2e8f0; margin: 0 0 16px 0; }
    .border-bottom { border-bottom: 1px solid #1e2130; padding-bottom: 12px; }

    /* Bulk Actions */
    .bulk-actions {
        display: flex; justify-content: space-between; align-items: flex-end;
        background: #0f1117; padding: 16px; border-radius: 6px; border: 1px solid #1e2130;
        margin-bottom: 16px; gap: 20px;
    }
    .bulk-form { display: flex; gap: 12px; flex: 1; }
    .bulk-buttons { display: flex; align-items: center; gap: 12px; }
    .selected-count { font-size: 13px; color: #94a3b8; font-weight: 600; }

    /* Forms & Inputs */
    .form-vertical { display: flex; flex-direction: column; gap: 12px; }
    .form-vertical label { font-size: 13px; color: #94a3b8; margin-bottom: -8px; }
    .input-field {
        background: #0f1117; border: 1px solid #1e2130; color: #e2e8f0;
        padding: 8px 12px; border-radius: 6px; font-size: 13px; outline: none;
    }
    .input-field:focus { border-color: #3b82f6; }

    /* Buttons */
    .btn {
        padding: 8px 16px; border: none; border-radius: 6px; font-size: 13px;
        font-weight: 600; cursor: pointer; transition: opacity 0.2s;
    }
    .btn:hover { opacity: 0.9; }
    .btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-primary { background: #2563eb; color: white; width: 100%; }
    .btn-danger { background: rgba(239,68,68,0.2); color: #ef4444; border: 1px solid rgba(239,68,68,0.3); }
    .btn-success { background: rgba(34,197,94,0.2); color: #22c55e; border: 1px solid rgba(34,197,94,0.3); }
    .btn-icon { background: transparent; border: none; cursor: pointer; padding: 4px 8px; border-radius: 4px; }
    .btn-icon:hover { background: rgba(255,255,255,0.05); }

    /* Tables */
    .table-container { overflow-x: auto; }
    .data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
    .data-table th {
        text-align: left; padding: 12px; color: #64748b; font-weight: 600;
        border-bottom: 1px solid #1e2130; background: #0f1117;
    }
    .data-table td { padding: 12px; border-bottom: 1px solid #1a1d2a; color: #cbd5e1; }
    .data-table tr:hover { background: #181b27; }
    .data-table tr.selected { background: rgba(59,130,246,0.05); }
    .empty-state { text-align: center; color: #64748b; padding: 32px !important; }

    .mono { font-family: monospace; }
    .badge { padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 700; }
    .badge-red { background: rgba(239,68,68,0.15); color: #ef4444; }
    .badge-orange { background: rgba(249,115,22,0.15); color: #f97316; }
    .badge-gray { background: #1e2130; color: #94a3b8; }
    .text-red { color: #ef4444; }
</style>
