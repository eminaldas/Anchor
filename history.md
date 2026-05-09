# AuditCore Self-Learning & Decision Log

Bu dosya, projenin geliştirilmesi sırasında alınan kritik mimari kararları, karşılaşılan sorunları, uygulanan çözümleri ve COSO/3LoD prensiplerine nasıl uyulduğunu kaydeder.

## [Kayıt 1] - Mimari Temellerin Atılması (Faz 1)
**Tarih:** 2026-04-11
**Konu:** Veritabanı Audit Loglama Yöntemi Seçimi
* **Karşılaşılan Durum:** `audit-logging-standard.md` gereğince logların "Append-Only" ve "Tamper-Evident" olması isteniyor. Bunu FastAPI (Backend) seviyesinde mi yoksa PostgreSQL (Database) seviyesinde mi yöneteceğimiz sorusuna en iyi yöntemi seçmem istendi.
* **Seçilen Çözüm:** **PostgreSQL Triggers & Functions** metodunu seçtim.
* **COSO/3LoD Nedeni:** Eğer loglamayı uygulama katmanına (FastAPI) bırakırsak, kodda oluşabilecek bir hata, zafiyet veya direkt veritabanına yapılan bir müdahale (Management Override) logları atlatabilir. PostgreSQL Trigger'ları ise tabloya gelen her `INSERT/UPDATE/DELETE` anında otomatik tetiklenir ve işlemi `audit_logs` tablosuna yazar. Bu, en yüksek seviye (İlke 16) bütünlük ve güvenceyi (Assurance) sağlar.

## [Kayıt 2] - State Management (Faz 1)
**Tarih:** 2026-04-11
**Konu:** Frontend Global State Yönetimi
* **Seçilen Çözüm:** Zustand.
* **Nedeni:** Redux'ın hantal yapısından kaçınmak, React Context API'nin aşırı render (re-render) sorunlarına takılmamak ve Fetch Wrapper ile entegrasyonu daha saf tutabilmek için kullanıcı kararıyla Zustand seçilmiştir.

## [Kayıt 3] - Suistimal Tespiti ve Browser Fingerprint (Faz 1)
**Tarih:** 2026-04-11
**Konu:** Tanımlanmayan cihazla kritik işlem girişimi.
* **Seçilen Çözüm:** Önleyici kontrol olarak doğrudan bloklama yapılmayacak; MFA/OTP istenecek. Risk Analisti tarafından anında "Güvenlik Alarmı" üretilecek ve işlem `audit_logs` tablosuna "Yeni Cihazla Kritik İşlem" etiketiyle kaydedilecek (İlke 8 ve 17).

## [Kay�t 4] - Bulgu Olu�turma Mod�l� Mimari Kararlar� (Faz 2)
**Tarih:** 2026-04-24
**Konu:** 4 C's Veri Do�rulama ve Sahiplik Atamas�
* **Se�ilen ��z�m:** Pydantic v2 modelleri (FastAPI seviyesinde) kullan�larak girdiler 5000 karaktere s�n�rland� (Data Extraction & Buffer Overflow g�venli�i). SQL taraf�nda ise standart String/Text tipi b�rak�larak gerekti�inde DB g��� yap�labilecek esneklik korundu. Sahiplik (Ownership) atamas�nda, form �zerinden do�rudan owner_id (1. Hat y�neticisi) se�ilmesi kararla�t�r�ld�. efore_state ve fter_state log verilerinde esneklik sa�lamak amac�yla veritaban�ndaki PostgreSQL JSONB destekli AuditLog kullan�ld�. Bu sayede nesnedeki eklemeler veritaban� �emas�n� bozmadan loglanabiliyor.
