# Finding Creation Module Plan

Bu doküman, AuditCore mimarisi içerisinde yer alan The Finding Creation (Bulgu Oluşturma) modülü için detaylı uygulama planını ve teknik seçimleri içermektedir.

## 1. Database Seeding (`backend/scripts/init_db.py`)
- `seed_coso_principles.json` dosyası okunarak `principles` tablosu `frameworks` tablosuna bağlı bir `framework_id` oluşturulduktan sonra doldurulacak.
- `org_units` tablosu için örnek 3 birim (IT, İK, Finans) seed edilecek.
- Güvenlik Önlemi: SQLAlchemy ORM kullanılacak, ham SQL (f-string) kesinlikle engellenecek.

## 2. Backend (FastAPI)
### Modeller (`app/models_finding.py`)
- `Finding` SQLAlchemy modeline 4 C (Condition, Criteria, Cause, Effect) için `String` (Text) tipinde sütunlar eklenecek.
### Şemalar (`app/api/findings_controller.py` veya `schemas.py`)
- Pydantic `FindingCreate` şeması ile veri doğrulama zorunlu tutulacak (`@field_validator` veya `Field(min_length=...)` ile).
- `condition, criteria, cause, effect` zorunlu alanlar olacak.
### Endpoint (`POST /api/findings`)
- Sadece `3rd Line` yetkiye (Denetçi) sahip ajanlar/kullanıcılar tarafından erişilebilecek.
- Bulgu oluşturulduğunda durumu `Open` atanacak.
- Seçilen `OrgUnit` kapsamına göre o departmanın bir çalışanına `owner_id` (1. Hat) yetkisiyle atanacak.
- COSO ve Audit standartlarına uygun JSON logu, `AuditLog` kullanılarak sisteme yazılacak.  

## 3. Frontend (React)
- `useFindingApi.ts` isimli güvenli (Pydantic validasyonlu API ile eşleşen) hook yaratılacak.
- `FindingCreateForm.tsx` bileşeni TailwindCSS kütüphanesi kullanılarak profesyonel ve modern (glassmorphism/vibrant öğeler içerik kirliliği yaratmayacak ölçüde) tasarlanacak.
- COSO ilkeleri ve `org_units` verisi için backend'den bir kere preload edilen `Select` komponentleri sunulacak.
- `dangerouslySetInnerHTML` veya güvensiz localStorage kullanımından tamamen kaçınılacak.

## 4. Teknik Kararlar (`history.md` kaydı)
- Pydantic v2 `Validation`'ların projenin Data Integrity prensibini nasıl sağladığı anlatılacak.
- Log yapısında JSONB kullanmanın, state diff alırken sağladığı avantajlar listelenecek.
