---
name: technical-security-audit
description: Teknik Güvenlik ve Kod Denetimi Skilli - FastAPI/React Güvenlik Checklistleri
role: Cybersecurity Auditor (3rd Line of Defense Support)
---

# Technical Security Audit — The Code Defense Skill

Bu skill, `audit-monitoring-agent` tarafından kullanılır ve kodun "Technical Integrity" (Teknik Bütünlük) denetimini yapar. COSO İlke 11'in kod seviyesindeki karşılığıdır.

## 1. FastAPI Security Layer (Backend)
- **Input Validation:** Zod yerine **Pydantic** kullanılmalıdır. Tüm inputlar `Field` tanımlarıyla (min/max length) kısıtlanmalıdır.
- **SQL Injection:** SQLAlchemy/SQLModel `session.exec()` dışındaki tüm string formatlı (f-string) SQL sorgularını "Kritik Hata" olarak işaretle.
- **CORS Management:** `allow_origins=["*"]` kullanımını yasakla; sadece React SPA domainine izin ver.
- **Rate Limiting:** `fastapi-limiter` veya Redis tabanlı koruma olup olmadığını doğrula (Brute Force koruması).

## 2. React Security Layer (Frontend)
- **Sensitive Data:** `localStorage` içinde JWT veya hassas veri saklanmasını yasakla; `httpOnly` cookie kullanımını zorunlu tut.
- **XSS Protection:** `dangerouslySetInnerHTML` kullanımını denetle ve mutlaka bir sanitizer (örn: DOMPurify) ile sarmala.
- **Fetch Usage:** React tarafında `fetch` kullanılırken global bir `apiClient` sarmalayıcısı üzerinden `CSRF-Token` eklenip eklenmediğini kontrol et.

## 3. Database & Docker Security (Postgres)
- **Principle of Least Privilege:** Uygulamanın veritabanına `superuser` yerine sadece belirli tablolara yetkili bir kullanıcı ile bağlanıp bağlanmadığını denetle.
- **Secret Management:** `.env` dosyalarının Docker imajı içine gömülmediğinden ve sistem çevre değişkenlerinden (env vars) okunduğundan emin ol.

## 4. 3LoD Entegrasyon Kuralı
- Bu skill ile bulunan teknik açıklar, `governance-workflow-specialist` tarafından otomatik olarak **"IT Departman Müdürü"**ne (1. Hat) atanır.
- Eğer açık 24 saat içinde kapatılmazsa, `audit-monitoring-agent` tarafından **"Kritik Bulgu"** olarak CTO'ya yükseltilir (Escalation).

## 5. Pre-Deployment Checklist (Mühendislik Kontrolü)
- [ ] Pydantic modelleri tüm inputları kapsıyor mu?
- [ ] SQLAlchemy sorguları parametrik mi?
- [ ] Redis yetki cache'lemesi aktif mi?
- [ ] Loglarda PII (Kişisel Veri) sızıntısı var mı?