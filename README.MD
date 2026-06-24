# Clinic Wise

## Getting started

Copy the environment file and update it for your local setup:

```
cp .env.example .env
```

Then run the application:

```
make run
```

The docker compose has a mysql service that can be used for the database.

## API overview

### Authentication

- `POST /auth/signup`
  - Patients can sign up directly.
  - Creates an active patient account.
- `POST /auth/signin`
- `POST /auth/invite`
  - Authenticated users can invite non-patient roles.
  - Invitee is created with `accepted = false`.
- `POST /auth/invites/:id/accept`
  - Accept an invite and set the password.
  - Marks the invite as accepted.
- `PATCH /auth/password`
  - Authenticated users can reset their password.

### Hospitals

- `POST /hospitals`
  - Admin only.
- `GET /hospitals/:id/stats`
  - Admin only.
  - Returns:
	- total appointments
	- active patients
	- prescription count

### Appointments

- `POST /appointments`
  - Authenticated users can create appointments.
- `GET /appointments?hospital_id=<id>`
  - Lists appointments for a hospital.
- `GET /appointments/user?page=<n>&limit=<m>`
  - Lists appointments for the authenticated user.
- `GET /appointments/:id`
  - Doctor or patient on the appointment only.
- `PATCH /appointments/:id/complete`
  - Doctor only.
  - Marks the appointment as `completed`.

### Notes

- `POST /notes`
- `PATCH /notes`
- `GET /notes/:id`
- `GET /notes/appointment/:id`

### Prescriptions

- `POST /prescriptions`
  - Doctor only.
  - Emits a `prescription_created` event.
- `PATCH /prescriptions/:id/dispatch`
  - Pharmacist only.
  - Cannot dispatch expired prescriptions.
  - Cannot dispatch prescriptions that are already `unavailable`.
  - Emits a `prescription_updated` event.
- `GET /prescriptions/:id`
  - Doctor or patient only.
- `GET /prescriptions/appointment/:id`
  - Doctor or patient only.

## Behavior notes

- Prescription event writes are best-effort: failures are logged and do not fail the request.
- If no queue writer is configured, the server uses a no-op prescription event writer.
- All role checks are enforced at the route layer.
