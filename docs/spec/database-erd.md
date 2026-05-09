# Database Entity-Relationship (ER) Logic

Bu döküman PostgreSQL modelinin teknik kurallarını içerir.

## 1. Core Schema

### Governance & Users
- `users`: id, email, pass_hash, full_name, is_active.
- `roles`: id, name, defense_line (1, 2, 3).
- `user_roles`: user_id, role_id.
- `frameworks`, `principles`, `objectives`: COSO Sabitleri.

### Risk & Control (2nd Line / 1st Line)
- `risks`: id, title, principle_id, risk_score.
- `controls`: id, risk_id, type (Preventive/Detective).

### Findings & Audit (3rd Line Hub)
- `findings`: id, principle_id, risk_id, status (Open, Resolved, Reviewed, Closed), owner_id, reviewer_id.
- `evidence`: id, finding_id, file_path, hash.

### Audit Trail (Security)
- `audit_logs`: trace_id, event_id, timestamp, user_id, user_role, action_type, schema_name, resource_id, before_state (JSONB), after_state (JSONB), coso_principle, fraud_risk_score.

## 2. PostgreSQL Triggers (The Enforcer)
**Amaç:** `audit_logs` tablosunu uygulama katmanından bağımsız olarak doldurmak.
Bir `findings` kaydı güncellendiğinde DB seviyesinde tetiklenen trigger, eski veriyi `before_state`, yeni veriyi `after_state` olarak JSONB formatında `audit_logs` tablosuna aktarır.

## 3. Hard Delete Koruması
`DELETE` komutu `findings`, `risks`, `audit_logs` tabloları için veritabanı policy'si ile tamamen yasaklanacak veya `is_deleted` flag'i üzerinden Rule/Trigger ile `UPDATE` komutuna çevrilecektir.
