# API Implementation Plan

## Phase 1: Context & Blueprint (Kapanıyor)
- [x] Tüm konseptin anlaşılması (COSO & 3LoD).
- [x] Temel mimari kararların alınması.
- [x] Dokümantasyon (`docs/spec`, `docs/plan`, `history.md`) kurulumu.

## Phase 2: Database & Infrastructure (Veritabanı ve Altyapı)
- [ ] Docker ve `docker-compose.yml` kurulumu (PosgreSQL, Redis).
- [ ] COSO verileri (İlkeler vs.) için initial seed scriptlerinin yazılması.
- [ ] PostgreSQL Audit Trigger fonksiyonlarının ve Tablo şemalarının (SQLAlchemy/Alembic) yapılandırılması.

## Phase 3: Backend Core & Security (API Temelleri)
- [ ] FastAPI projesinin başlatılması (Pydantic v2 ile).
- [ ] Auth sistemi (JWT, Redis session/fingerprint caching).
- [ ] Custom Security Dependency'ler (CORS, Rate Limiting, MFA Trigger mekanizması).
- [ ] 3LoD yetki matrisi middleware'in yazılması.

## Phase 4: Frontend Architecture (Arayüz Temelleri)
- [ ] React/Vite projesi kurulumu.
- [ ] TailwindCSS Atomic bileşen sisteminin (Design System) oturtulması.
- [ ] Axios yasağına uygun Fetch API Wrapper yazılması (Interceptor mantığı ile yetki sınırlandırması).
- [ ] Zustand mağazasının (Store) session stateleri için kurulması.

## Phase 5: Business Engine & Workflows (İş Akışları)
- [ ] Bulguların (Findings) State Machine uç noktalarının yazılması.
- [ ] İzleme ve Güvenlik Alarmları dashboard'unun kodlanması.
- [ ] "Yeni cihaz" MFA workflow testleri.
