# System Blueprint & Architecture

## 1. Tech Stack
- **Backend:** FastAPI (Async), Pydantic v2 (Validation), SQLAlchemy (ORM).
- **Database:** PostgreSQL (Relational Data & JSONB for State), Redis (Caching, Pub/Sub, Rate Limiting).
- **Frontend:** React (Vite), Tailwind CSS, Zustand (State Management).
- **Network / API:** Fetch API with Custom Wrapper (No Axios), RESTful JSON endpoints.
- **Infrastructure:** Docker & Docker Compose (Microservices structure).

## 2. 3LoD Access Control Layer
Role-Based Access Control (RBAC) ile yönetilecektir:
- **1st Line (Operasyonel Sahipler):** Kendi departmanlarındaki bulguları yönetir, kanıt yükler, kontrolleri uygular.
- **2nd Line (Risk & Uyum):** İlk hattı izler, risk skorlaması yapar, MFA alarmlarını inceler.
- **3rd Line (Denetim):** Sadece salt-okunur (read-only) ve "Kapatma (Close Finding)" onayı yetkilerine sahiptir. Operasyonel değişiklik yapamaz, kontrol tasarlayamaz.

## 3. Security Fundamentals (İlke 11)
- **CORS:** Sadece React SPA domainine izin verilecek.
- **Rate Limiting:** Redis tabanlı, özellikle kritik endpointlerde (Örn: MFA, Login) brute-force koruması.
- **Data Retention & Soft Deletion:** `is_deleted` bayrağı ile soft-delete zorunlu; hard delete veritabanı seviyesinde engellenecek.
- **Tamper-Evident Logging:** PostgreSQL Trigger tabanlı immutable (değiştirilemez) loglama.
