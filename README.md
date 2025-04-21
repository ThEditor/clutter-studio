# Clutter Analytics

A lightweight web analytics solution with three main components: a tracking script, collection backend, and analytics dashboard.

## Architecture

```
├── frame/      # Frontend dashboard (Next.js)
├── ink/        # Tracking script (JavaScript)
├── paper/      # Collection backend (Go)
└── studio/     # Dashboard backend (Go)
```

You can view the system design [here](https://www.figma.com/board/iSkI8Wf4Bg2ObJvqozN4EJ/Clutter?node-id=0-1&t=cEBS1bcxrVjqWn0B-1).

### Components

#### Frame (Frontend Dashboard)
- Next.js web application that visualizes analytics data
- Features:
  - Site management dashboard
  - Simple analytics data visualization with charts and metrics
- Checkout the github repository [here](https://github.com/ThEditor/clutter-frame)

#### Ink (Tracking Script) 
- Lightweight JavaScript snippet that website owners add to their sites
- Collects basic analytics data:
  - Page views
  - User agent info
  - Referrer data
- Makes HTTP requests to Paper backend to store events
- Checkout the github repository [here](https://github.com/ThEditor/clutter-ink)
- Configuration via `window.clutterConfig`:
```js
window.clutterConfig = {
  siteId: "your-site-id"
}
```

#### Paper (Collection Backend)
- Go service that accepts analytics events from Ink
- Features:
  - CORS support for cross-origin requests
  - Request validation
  - ClickHouse database for event storage
- API Endpoints:
  - `POST /api/event` - Records analytics events
- Checkout the github repository [here](https://github.com/ThEditor/clutter-paper)

#### Studio (Dashboard Backend)
- Go service that powers the Frame frontend
- Features:
  - User authentication with JWT
  - Site management (CRUD operations)
  - Analytics data access from ClickHouse
- Key APIs:
  - `/auth` - User registration/login
  - `/sites` - Site management 
  - `/sites/{id}/analytics` - Analytics data retrieval
- Checkout the github repository [here](https://github.com/ThEditor/clutter-studio)

### Data Storage

- PostgreSQL - User and site data (Studio)
- ClickHouse - Analytics events data (Paper)
- Redis - Stores data that needs to be quickly communicated between Studio and Paper.

### Development

Requirements:
- Go 1.24+
- Node.js
- PostgreSQL
- ClickHouse
- Redis

Environment variables:
```sh
# Studio
DATABASE_URL=postgres://user:pass@localhost:5432/db
CLICKHOUSE_URL=clickhouse://default:@localhost:9000/clutter
PORT=8081
JWT_SECRET=secret

# Paper
DATABASE_URL=clickhouse://default:@localhost:9000/clutter
REDIS_URL=redis://user:pass@localhost:6379
POSTGRES_URL=postgres://user:pass@localhost:5432/db
PORT=8080
```

### Project Status

This is a basic analytics implementation with core features like:
- Site tracking via JavaScript snippet
- Event collection API
- User authentication
- Site management
- Basic analytics viewing

This project is still under progress.
