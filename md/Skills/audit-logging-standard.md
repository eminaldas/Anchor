---
name: audit-logging-standard
description: Evrensel Denetim İzi (Audit Trail) ve Loglama Standardı - COSO & 3LoD Uyumlu
scope: System-wide Logging Policy (Backend, Database, Admin)
---

# Audit Logging Standard — The "Unforgettable" Record

Bu döküman, sistemdeki her türlü veri hareketinin ve yönetsel kararın "inkar edilemez" (non-repudiation) bir kaydını tutmak için gereken teknik detayları tanımlar. [cite_start]COSO İlke 13 uyarınca bilginin **tam, doğru ve erişilebilir** olmasını hedefler.

## 1. Log Yapısı (The Data Schema)
Her bir log kaydı (JSON formatında) şu yedi ana bloktan oluşmalıdır:

### A. Metadata (Kimlik)
- `trace_id`: Mikroservisler arası işlem takibi için benzersiz ID.
- `timestamp`: UTC formatında mikro saniye hassasiyetinde zaman mührü.
- `event_id`: Logun kendi benzersiz anahtarı.

### [cite_start]B. Actor (Kim yaptı?) [cite: 108]
- `user_id`: İşlemi yapan kullanıcının ID'si.
- `user_role`: İşlem anındaki savunma hattı (1st, 2nd, or 3rd Line).
- `session_id`: Oturum bazlı takip için.
- `source_ip`: İşlemin geldiği IP adresi ve cihaz bilgisi.

### C. Action (Ne yaptı?)
- `action_type`: (CREATE, READ, UPDATE, DELETE, LOGIN, EXPORT, OVERRIDE).
- `endpoint`: Çağrılan API yolu.
- `status`: (SUCCESS, FAILURE, DENIED).

### D. Resource (Neye yaptı?)
- `resource_type`: (Finding, Risk, User, Config, Wallet).
- `resource_id`: Etkilenen kaydın ID'si.
- `schema_name`: Veritabanı tablo adı.

### [cite_start]E. Data Delta (Değişim Ne?) [cite: 237]
- `before_state`: Verinin işlemden önceki JSON hali.
- `after_state`: Verinin işlemden sonraki JSON hali.
- `changes`: Sadece değişen alanların listesi (Diff).

### F. Compliance & Risk Tags
- `coso_principle`: İlgili COSO ilkesi (1-17).
- `compliance_tag`: (GDPR, SOC2, KVKK).
- [cite_start]`fraud_risk_score`: Analist agent tarafından atanan risk puanı[cite: 473].

### G. Integrity (Güvenlik)
- `previous_hash`: Bir önceki logun hash değeri (Blokzincir mantığıyla tamper-proof zincir).
- `signature`: Sistemin bu logu onayladığını gösteren dijital imza.

## 2. Loglama Kategorileri ve Öncelikler
| Kategori | Örnek Olay | Saklama Süresi |
| :--- | :--- | :--- |
| **Sistem** | Başlatma, kapatma, konfigürasyon değişikliği. | 1 Yıl |
| **Güvenlik** | Başarısız giriş, yetki aşımı, şifre değişikliği. | 3 Yıl |
| **Denetim** | Bulgu silme, risk skoru değiştirme, rapor export. | 6 Yıl (HIPAA uyumlu) |
| **Yönetsel** | Management Override (Kontrolü baypas etme). | Sonsuz / Arşiv |

## 3. Asla Loglanmaması Gerekenler (Forbidden Data)
- Kullanıcı şifreleri (hashlenmiş olsa dahi).
- Tam kredi kartı numaraları veya CVC kodları.
- Ham PII verileri (Örn: Logda isim yerine `user_id` kullanılmalı).
- API Keyler ve Private Keyler.

## 4. Güvenlik ve Bütünlük Kuralları
1. **Append-Only:** Loglar asla güncellenemez veya silinemez.
2. **Centralization:** Loglar ana veritabanından ayrı, salt-okunur (WORM) bir depolama birimine anlık olarak akmalıdır.
3. [cite_start]**Tamper-Evident:** Log zincirindeki herhangi bir hash bozulması durumunda `audit-monitoring-agent` derhal Yönetim Kurulu'na "Kritik Alarm" üretmelidir.

## 5. Agentlar İçin Log Sorgulama Rehberi
- **Risk Analyst:** "Hangi kullanıcı mesai saatleri dışında 10'dan fazla kritik risk raporu indirdi?"
- **Audit Agent:** "Yönetici, kendi yetkisini kullanarak hangi kapatılmış bulguları tekrar açtı?"