# Guía de Integración: SparkBIGS CRM API + n8n

## Índice
1. [Cómo obtener tu API Key](#1-cómo-obtener-tu-api-key)
2. [Configurar la autenticación en n8n](#2-configurar-la-autenticación-en-n8n)
3. [Nodo HTTP Request — configuración base](#3-nodo-http-request--configuración-base)
4. [Empresas (Companies)](#4-empresas-companies)
5. [Contactos (Contacts)](#5-contactos-contacts)
6. [Reuniones (Meetings)](#6-reuniones-meetings)
7. [Suscripciones (Subscriptions)](#7-suscripciones-subscriptions)
8. [Manejo de errores en n8n](#8-manejo-de-errores-en-n8n)
9. [Ejemplos de flujos completos](#9-ejemplos-de-flujos-completos)

---

## 1. Cómo obtener tu API Key

Antes de conectar n8n necesitas generar una API Key desde el CRM. La clave **solo se muestra una vez**, guárdala en un lugar seguro.

### Paso 1 — Inicia sesión en el CRM y obtén tu JWT

```http
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "tu@email.com",
  "password": "tu_contraseña"
}
```

Respuesta:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "uuid-del-refresh-token",
    "user": { "id": 1, "name": "Tu Nombre", "role": "admin" }
  }
}
```

### Paso 2 — Crea una API Key (usando el JWT del paso anterior)

```http
POST http://localhost:8080/api/v1/apikeys/
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "name": "n8n-integration",
  "scopes": "webhooks"
}
```

Respuesta:
```json
{
  "success": true,
  "data": {
    "plaintext_key": "spk_a1b2c3d4e5f6...",
    "key": {
      "id": 1,
      "name": "n8n-integration",
      "key_prefix": "spk_a1b2",
      "scopes": "webhooks",
      "is_active": true,
      "created_at": "2026-04-02T10:00:00Z"
    }
  }
}
```

> ⚠️ **IMPORTANTE:** El campo `plaintext_key` solo aparece en esta respuesta. Cópialo ahora — no hay forma de recuperarlo después.

---

## 2. Configurar la autenticación en n8n

En n8n, crea una **Credential** de tipo **Header Auth** para reutilizarla en todos los nodos:

1. Ve a **Settings → Credentials → New Credential**
2. Tipo: **Header Auth**
3. Configura:
   - **Name:** `SparkBIGS CRM API Key`
   - **Header Name:** `X-API-Key`
   - **Header Value:** `spk_a1b2c3d4e5f6...` (tu clave del paso anterior)
4. Guarda la credencial.

---

## 3. Nodo HTTP Request — configuración base

Todos los nodos **HTTP Request** de n8n para esta API comparten esta configuración:

| Campo | Valor |
|-------|-------|
| **Method** | GET / POST / PUT / DELETE según la operación |
| **URL** | `http://localhost:8080/webhooks/v1/{recurso}` |
| **Authentication** | Header Auth → `SparkBIGS CRM API Key` |
| **Send Headers** | `Content-Type: application/json` (solo en POST/PUT) |
| **Response Format** | JSON |

> Si el CRM está en producción, reemplaza `http://localhost:8080` por tu dominio real.

---

## 4. Empresas (Companies)

Base URL: `http://localhost:8080/webhooks/v1/companies`

### 4.1 Listar empresas

| Campo nodo | Valor |
|------------|-------|
| Method | `GET` |
| URL | `http://localhost:8080/webhooks/v1/companies` |
| Query Parameters | `offset=0`, `limit=20` |

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "Acme Corp",
        "sector": "Tecnología",
        "status": "active",
        "website": "https://acme.com",
        "phone": "+34 600 000 000",
        "address": "Calle Mayor 1, Madrid",
        "relation_start_date": "2025-01-15T00:00:00Z",
        "user_id": 1
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 20
  }
}
```

En n8n, accede a los datos con la expresión: `{{ $json.data.list }}`

---

### 4.2 Obtener una empresa por ID

| Campo nodo | Valor |
|------------|-------|
| Method | `GET` |
| URL | `http://localhost:8080/webhooks/v1/companies/{{ $json.company_id }}` |

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "company": {
      "id": 1,
      "name": "Acme Corp",
      "sector": "Tecnología",
      "status": "active"
    }
  }
}
```

---

### 4.3 Crear empresa

| Campo nodo | Valor |
|------------|-------|
| Method | `POST` |
| URL | `http://localhost:8080/webhooks/v1/companies` |
| Body Type | JSON |

**Body:**
```json
{
  "name": "Nueva Empresa S.L.",
  "sector": "Consultoría",
  "status": "prospect",
  "website": "https://nuevaempresa.com",
  "phone": "+34 611 222 333",
  "address": "Av. Diagonal 123, Barcelona",
  "relation_start_date": "2026-04-01"
}
```

Campos obligatorios: `name`  
Valores de `status`: `prospect` | `active` | `inactive`

**Respuesta (201):**
```json
{
  "success": true,
  "data": {
    "company": { "id": 5, "name": "Nueva Empresa S.L.", ... }
  }
}
```

---

### 4.4 Actualizar empresa

| Campo nodo | Valor |
|------------|-------|
| Method | `PUT` |
| URL | `http://localhost:8080/webhooks/v1/companies/{{ $json.id }}` |
| Body Type | JSON |

**Body** (solo los campos que quieras cambiar):
```json
{
  "status": "active",
  "phone": "+34 699 888 777"
}
```

---

### 4.5 Eliminar empresa

| Campo nodo | Valor |
|------------|-------|
| Method | `DELETE` |
| URL | `http://localhost:8080/webhooks/v1/companies/{{ $json.id }}` |

**Respuesta:**
```json
{
  "success": true,
  "data": { "deleted": true }
}
```

---

## 5. Contactos (Contacts)

Base URL: `http://localhost:8080/webhooks/v1/contacts`

### 5.1 Listar contactos

```
GET /webhooks/v1/contacts?offset=0&limit=20
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "Juan García",
        "email": "juan@acme.com",
        "phone": "+34 600 111 222",
        "position": "Director de Compras",
        "status": "active",
        "company_id": 1,
        "user_id": 1
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 20
  }
}
```

### 5.2 Obtener contacto por ID

```
GET /webhooks/v1/contacts/{{ $json.contact_id }}
```

### 5.3 Crear contacto

```
POST /webhooks/v1/contacts
```

**Body:**
```json
{
  "name": "María López",
  "email": "maria@empresa.com",
  "phone": "+34 611 333 444",
  "position": "CEO",
  "status": "active",
  "company_id": 3
}
```

Campos obligatorios: `name`  
`company_id` es opcional (puede omitirse si el contacto no pertenece a ninguna empresa)  
Valores de `status`: `active` | `inactive` | `lead`

### 5.4 Actualizar contacto

```
PUT /webhooks/v1/contacts/{{ $json.id }}
```

**Body** (campos a modificar):
```json
{
  "position": "CTO",
  "email": "nuevo@email.com"
}
```

### 5.5 Eliminar contacto

```
DELETE /webhooks/v1/contacts/{{ $json.id }}
```

---

## 6. Reuniones (Meetings)

Base URL: `http://localhost:8080/webhooks/v1/meetings`

### 6.1 Listar reuniones

```
GET /webhooks/v1/meetings?offset=0&limit=20
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "title": "Demo producto",
        "company_id": 1,
        "contact_id": 2,
        "start_at": "2026-04-10T10:00:00Z",
        "duration_min": 60,
        "status": "scheduled",
        "notes": "Llevar presentación",
        "summary": "El cliente mostró interés en el plan Pro. Próximo paso: enviar propuesta."
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 20
  }
}
```

### 6.2 Obtener reunión por ID

```
GET /webhooks/v1/meetings/{{ $json.meeting_id }}
```

### 6.3 Crear reunión

```
POST /webhooks/v1/meetings
```

**Body:**
```json
{
  "title": "Reunión de seguimiento",
  "company_id": 1,
  "contact_id": 2,
  "start_at": "2026-04-15T09:30:00Z",
  "duration_min": 45,
  "status": "scheduled",
  "notes": "Revisar propuesta comercial",
  "summary": ""
}
```

Campos obligatorios: `title`, `company_id`, `start_at`  
`contact_id` es opcional  
`start_at` debe estar en formato **RFC3339**: `YYYY-MM-DDTHH:MM:SSZ`  
`duration_min` por defecto es `60`  
Valores de `status`: `scheduled` | `completed` | `cancelled`  
`summary` es opcional — campo de texto libre para el resumen post-reunión (acuerdos, próximos pasos, etc.)

### 6.4 Actualizar reunión

```
PUT /webhooks/v1/meetings/{{ $json.id }}
```

**Body:**
```json
{
  "status": "completed",
  "notes": "Reunión realizada. Cliente interesado en plan Pro.",
  "summary": "Acordado enviar propuesta antes del viernes. Juan confirmó presupuesto de 5.000€. Próxima reunión: 22 abril."
}
```

> El campo `summary` es ideal para actualizarlo desde n8n justo después de la reunión, por ejemplo al recibir un webhook de Notion, Google Meet o un formulario de feedback del equipo.

### 6.5 Eliminar reunión

```
DELETE /webhooks/v1/meetings/{{ $json.id }}
```

---

## 7. Suscripciones (Subscriptions)

Base URL: `http://localhost:8080/webhooks/v1/subscriptions`

### 7.1 Listar suscripciones

```
GET /webhooks/v1/subscriptions?offset=0&limit=20
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "Plan Pro Anual",
        "company_id": 1,
        "plan_type": "pro",
        "status": "active",
        "amount": 1200.00,
        "currency": "EUR",
        "billing_cycle": "annual",
        "start_date": "2026-01-01T00:00:00Z",
        "renewal_date": "2027-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 20
  }
}
```

### 7.2 Obtener suscripción por ID

```
GET /webhooks/v1/subscriptions/{{ $json.subscription_id }}
```

### 7.3 Crear suscripción

```
POST /webhooks/v1/subscriptions
```

**Body:**
```json
{
  "name": "Plan Básico Mensual",
  "company_id": 2,
  "plan_type": "basic",
  "status": "active",
  "amount": 99.00,
  "currency": "EUR",
  "billing_cycle": "monthly",
  "start_date": "2026-04-01",
  "renewal_date": "2026-05-01",
  "notes": "Primer mes bonificado"
}
```

Campos obligatorios: `name`, `company_id`, `start_date`  
`start_date` y `renewal_date` en formato `YYYY-MM-DD`  
Valores de `billing_cycle`: `monthly` | `quarterly` | `annual` | `one_time`  
Valores de `status`: `active` | `trial` | `expired` | `cancelled`  
`currency` por defecto: `EUR`

### 7.4 Actualizar suscripción

```
PUT /webhooks/v1/subscriptions/{{ $json.id }}
```

**Body:**
```json
{
  "status": "cancelled",
  "notes": "Cliente canceló por presupuesto"
}
```

### 7.5 Eliminar suscripción

```
DELETE /webhooks/v1/subscriptions/{{ $json.id }}
```

---

## 8. Manejo de errores en n8n

Todas las respuestas de error siguen este formato:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Empresa no encontrada"
  }
}
```

| Código HTTP | Code en JSON | Significado |
|-------------|--------------|-------------|
| 400 | `INVALID_BODY` | JSON malformado en el body |
| 400 | `VALIDATION_FAILED` | Campo obligatorio faltante |
| 400 | `INVALID_DATE` | Formato de fecha incorrecto |
| 401 | `MISSING_API_KEY` | No se envió el header `X-API-Key` |
| 401 | `INVALID_API_KEY` | La clave no existe o está revocada |
| 401 | `API_KEY_EXPIRED` | La clave ha expirado |
| 404 | `NOT_FOUND` | El recurso no existe o no te pertenece |
| 500 | `INTERNAL_ERROR` | Error interno del servidor |

### Configurar manejo de errores en n8n

En el nodo HTTP Request activa la opción **"Continue on Fail"** y añade un nodo **IF** después:

```
Condición: {{ $json.success }} === false
→ Rama TRUE (error): nodo para notificar o registrar el error
→ Rama FALSE (éxito): continúa el flujo normal
```

---

## 9. Ejemplos de flujos completos

### Flujo 1: Registrar nuevo lead desde formulario externo

```
[Webhook Trigger (formulario)]
    ↓
[HTTP Request — POST /webhooks/v1/companies]
    Body: { "name": "{{ $json.company }}", "status": "prospect" }
    ↓
[HTTP Request — POST /webhooks/v1/contacts]
    Body: {
      "name": "{{ $('Webhook Trigger').item.json.name }}",
      "email": "{{ $('Webhook Trigger').item.json.email }}",
      "company_id": {{ $json.data.company.id }}
    }
    ↓
[Send Email / Slack — Notificar al equipo]
```

---

### Flujo 2: Actualizar estado de empresa al recibir pago

```
[Webhook Trigger (Stripe/pasarela)]
    ↓
[IF: $json.type === "payment.succeeded"]
    ↓ TRUE
[HTTP Request — GET /webhooks/v1/companies?offset=0&limit=1]
    (buscar empresa por nombre u otro criterio externo)
    ↓
[HTTP Request — PUT /webhooks/v1/companies/{{ $json.data.list[0].id }}]
    Body: { "status": "active" }
    ↓
[HTTP Request — POST /webhooks/v1/subscriptions]
    Body: {
      "name": "{{ $('Webhook Trigger').item.json.plan }}",
      "company_id": {{ $json.data.company.id }},
      "amount": {{ $('Webhook Trigger').item.json.amount }},
      "start_date": "{{ $now.format('yyyy-MM-dd') }}"
    }
```

---

### Flujo 3: Informe diario de reuniones del día

```
[Schedule Trigger — cada día a las 08:00]
    ↓
[HTTP Request — GET /webhooks/v1/meetings?offset=0&limit=50]
    ↓
[Code Node — filtrar reuniones de hoy]
    const hoy = new Date().toISOString().split('T')[0];
    return $input.all().filter(item =>
      item.json.data.list.some(m => m.start_at.startsWith(hoy))
    );
    ↓
[Send Email / Slack — Enviar resumen]
```

---

### Flujo 4: Guardar resumen automáticamente tras una reunión

Útil para capturar el resumen desde un formulario de Google Forms, Typeform, Notion, etc.

```
[Webhook Trigger — formulario post-reunión]
    (campos: meeting_id, summary_text)
    ↓
[HTTP Request — PUT /webhooks/v1/meetings/{{ $json.meeting_id }}]
    Method: PUT
    Header: X-API-Key: spk_...
    Body: {
      "status": "completed",
      "summary": "{{ $json.summary_text }}"
    }
    ↓
[IF: $json.success === true]
    → TRUE: [Slack — "Resumen guardado para reunión {{ $json.data.meeting.title }}"]
    → FALSE: [Slack — "Error al guardar resumen: {{ $json.error.message }}"]
```

---

### Flujo 5: Sincronización con CRM externo (importar contactos)

```
[HTTP Request — GET API externa]
    ↓
[Loop Over Items]
    ↓ (por cada contacto externo)
[HTTP Request — POST /webhooks/v1/contacts]
    Body: {
      "name": "{{ $json.full_name }}",
      "email": "{{ $json.email }}",
      "phone": "{{ $json.phone_number }}",
      "status": "active"
    }
    ↓
[IF: $json.success === false]
    → Registrar error en Google Sheets / base de datos
```

---

## Referencia rápida de endpoints

| Recurso | Listar | Obtener | Crear | Actualizar | Eliminar |
|---------|--------|---------|-------|------------|---------|
| Empresas | `GET /companies` | `GET /companies/:id` | `POST /companies` | `PUT /companies/:id` | `DELETE /companies/:id` |
| Contactos | `GET /contacts` | `GET /contacts/:id` | `POST /contacts` | `PUT /contacts/:id` | `DELETE /contacts/:id` |
| Reuniones | `GET /meetings` | `GET /meetings/:id` | `POST /meetings` | `PUT /meetings/:id` | `DELETE /meetings/:id` |
| Suscripciones | `GET /subscriptions` | `GET /subscriptions/:id` | `POST /subscriptions` | `PUT /subscriptions/:id` | `DELETE /subscriptions/:id` |

**Base URL:** `http://localhost:8080/webhooks/v1`  
**Autenticación:** Header `X-API-Key: spk_...` en todas las peticiones  
**Paginación:** Query params `?offset=0&limit=20` en todos los endpoints de listado
