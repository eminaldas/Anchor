---
name: database-schema-infrastructure
description: Merkezi Veritabanı Mimarisi ve İlişkisel Model - COSO, 3LoD ve Audit Core Entegrasyonu
role: Lead Database Architect & System Designer
---

# Database Schema & Infrastructure — The System Backbone

Bu döküman, projenin tüm bileşenlerini (Hedefler, Riskler, Kontroller, Bulgular ve Loglar) birbirine bağlayan ilişkisel modeli tanımlar. PostgreSQL üzerinde "Data Integrity" (Veri Bütünlüğü) ve "Traceability" (İzlenebilirlik) odaklı bir yapı kurar.

## 1. Katmanlı Veri Modeli

### A. Yönetişim Katmanı (Governance Layer)
COSO Küpü'nün sabitlerini ve organizasyon şemasını tutar.
- **`frameworks`**: (id, name, version) -> Örn: "COSO 2013", "ISO 27001".
- **`principles`**: (id, framework_id, code, title, description) -> 17 İlke burada tanımlanır.
- **`objectives`**: (id, name) -> Operations, Reporting, Compliance.
- **`org_units`**: (id, name, parent_id) -> Departmanlar ve hiyerarşi.

### B. Kimlik ve Yetki Katmanı (3LoD Identity Layer)
3 Savunma Hattı rollerini ve yetki sınırlarını yönetir.
- **`users`**: (id, email, full_name, is_active, browser_fingerprint).
- **`roles`**: (id, name, defense_line) -> 1st, 2nd, 3rd Line.
- **`permissions`**: (id, action, resource) -> (Read, Write, Delete, Audit_Approve).
- **`user_unit_assignments`**: (user_id, unit_id) -> Hangi kullanıcı hangi birimden sorumlu?

### C. Risk ve Kontrol Katmanı (Risk & Control Layer)
Tehditleri ve onlara karşı kurulan bariyerleri tutar.
- **`risks`**: (id, title, description, principle_id, objective_id, owner_unit_id).
- **`risk_assessments`**: (id, risk_id, likelihood, impact, score, analyst_id).
- **`controls`**: (id, risk_id, control_type, frequency, description) -> (Preventive, Detective).

### D. Uygulama Katmanı (Audit & Findings Layer) - **PROJENİN KALBİ**
Saha çalışmalarını ve bulunan hataları yönetir.
- **`audits`**: (id, title, status, start_date, end_date, lead_auditor_id).
- **`findings`**: (id, audit_id, principle_id, risk_id, control_id, severity, status).
    - *İlişki:* Bir bulgu mutlaka bir ilke, bir risk ve (varsa) çalışmayan bir kontrolle bağlıdır.
- **`corrective_actions`**: (id, finding_id, assignee_id, due_date, status, resolution_details).
- **`evidence`**: (id, finding_id, action_id, file_path, uploaded_at, uploader_id).

### E. İzleme ve Log Katmanı (Audit Trail Layer)
Daha önce tanımlanan `audit-logging-standard.md` dosyasının fiziksel karşılığı.
- **`audit_logs`**: (id, trace_id, actor_id, action, resource_type, before_state, after_state, hash_chain, ip_address, fingerprint).

## 2. Kritik Tablo İlişkileri (ERD Logic)
- **Finding -> Principle (Many-to-One):** Her bulgu hangi COSO ilkesinin (1-17) ihlal edildiğini bilmek zorundadır.
- **Finding -> 3LoD Flow:** - `finding.created_by` -> 3rd Line (Denetçi).
    - `finding.owner` -> 1st Line (İş Sahibi).
    - `finding.reviewer` -> 2nd Line (Risk/Uyum).
- **Risk -> Objective:** Her risk, COSO'nun 3 ana hedefinden (Operations, Reporting, Compliance) en az birini tehdit etmelidir.

## 3. Altyapı ve Performans Kuralları
- **PostgreSQL:** Ana veri deposu. `JSONB` veri tipi `before_state` ve `after_state` logları için zorunludur.
- **Redis:** `user_permissions` ve `session_data` burada cache'lenir. Anlık bildirimler (WebSocket) için `Pub/Sub` mekanizması kullanılır.
- **Hard Constraints:** `findings` tablosundan veri silinmesi (Hard Delete) veritabanı seviyesinde yasaktır; sadece `is_deleted` flag'i kullanılabilir ve bu işlem loglanmalıdır.

## 4. "Pro" Özellikler (Advanced Intelligence)
- **Browser Fingerprinting:** Kullanıcı girişlerinde `browser_fingerprint` kontrol edilerek, kritik işlemlerin (örn: bulgu kapatma) her zaman aynı cihazdan yapılıp yapılmadığı denetlenir (Fraud Detection).
- **State Machine Integration:** `findings.status` değişimi (Open -> In Progress -> Closed) sadece `governance-workflow-specialist` agent'ının onayladığı yetki matrisine göre yapılabilir.