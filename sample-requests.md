# API Sample Requests

Base URL: `http://localhost:9040`

---

## Health Check

```bash
curl -X GET http://localhost:9040/health
```

**Response:**
```json
{
  "success": true,
  "message": "ok"
}
```

---

## Add User

Creates a new user along with their contact and address records.  
The username is auto-generated as `fname + mname + lname` (no separators).

```bash
curl -X POST http://localhost:9040/addusers \
  -H "Content-Type: application/json" \
  -d '{
    "fullname": "Juan Dela Cruz",
    "fname": "Juan",
    "lname": "Cruz",
    "mname": "Dela",
    "staffid": "STF12345",
    "contactnumber": "+639171234567",
    "email": "juan.cruz@example.com",
    "telephonenumber": "0287654321",
    "mainaddress": "123 Mabini Street, Manila",
    "secondaryaddress": "Unit 5, Sunshine Apartments",
    "lastaddress": "Old Residence, Quezon City",
    "birthdate": "1990-05-15T00:00:00Z",
    "userroleid": 1,
    "userpassword": "SecurePass123!"
  }'
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "user added successfully"
}
```

---

## Update User

Updates an existing user. All fields are sent but only the ones supported by the function are applied.  
`userpassword` is accepted in the payload but is ignored by the update function (password changes are not supported).

```bash
curl -X PUT http://localhost:9040/updateuser/1 \
  -H "Content-Type: application/json" \
  -d '{
    "fullname": "Juan Dela Cruz",
    "fname": "Juan",
    "lname": "Cruz",
    "mname": "Dela",
    "staffid": "STF12345",
    "contactnumber": "+639181234567",
    "email": "juan.dela.cruz@newdomain.com",
    "telephonenumber": "0287654321",
    "mainaddress": "456 Rizal Avenue, Makati",
    "secondaryaddress": "Unit 5, Sunshine Apartments",
    "lastaddress": "Old Residence, Quezon City",
    "birthdate": "1990-05-15T00:00:00Z",
    "userroleid": 1,
    "userpassword": "SecurePass123!"
  }'
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "user updated successfully"
}
```

---

## Delete User

Deletes the user and cascades to remove their credentials, contacts, and address records.

```bash
curl -X DELETE http://localhost:9040/deleteuser/1
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "user deleted successfully"
}
```

---

## Request Body Fields

| Field              | Type   | JSON Key           | Add | Update | Description                          |
|--------------------|--------|--------------------|-----|--------|--------------------------------------|
| `FullName`         | string | `fullname`         | ✓   | ✓      | Complete full name                   |
| `FirstName`        | string | `fname`            | ✓   | ✓      | First name                           |
| `LastName`         | string | `lname`            | ✓   | ✓      | Last name                            |
| `MiddleName`       | string | `mname`            | ✓   | ✓      | Middle name                          |
| `StaffID`          | string | `staffid`          | ✓   | ✓      | Employee / staff ID                  |
| `ContactNumber`    | string | `contactnumber`    | ✓   | ✓      | Mobile/phone number                  |
| `BirthDate`        | string | `birthdate`        | ✓   | ✓      | Birthdate (ISO 8601)                 |
| `Email`            | string | `email`            | ✓   | ✓      | Email address                        |
| `TelephoneNumber`  | string | `telephonenumber`  | ✓   | ✓      | Landline number                      |
| `MainAddress`      | string | `mainaddress`      | ✓   | ✓      | Primary address                      |
| `SecondaryAddress` | string | `secondaryaddress` | ✓   | ✓      | Secondary address                    |
| `LastAddress`      | string | `lastaddress`      | ✓   | ✓      | Previous address                     |
| `UserRoleID`       | int    | `userroleid`       | ✓   | ✓      | Role ID (must exist in UserAccountType) |
| `Password`         | string | `userpassword`     | ✓   | ✗      | Account password                     |

**Add User** — `fname`, `lname`, and `userpassword` are effectively required (the username is built from first/middle/last name, and password is stored in `UserCredentials`).

**Update User** — all fields are sent in the payload. `userpassword` is accepted but ignored by the backend. `userroleid` must reference an existing role in `UserAccountType`.

---

## Workflow

```
1. Seed role:             INSERT INTO UserAccountType (RoleDesc, Status) VALUES ('Admin', 1);
2. Add User (ID 1):       POST /addusers
3. Update User (ID 1):    PUT  /updateuser/1
4. Delete User (ID 1):    DELETE /deleteuser/1
```

---

## Notes

- All requests use `Content-Type: application/json`.
- All responses follow the format: `{ "success": bool, "message": string, "data": ... }`.
- The database must have at least one role in `UserAccountType` before adding users. If the init script didn't seed it, run:

```bash
docker exec -i kbpos-postgres psql -U myuser -d postgres -c "INSERT INTO UserAccountType (RoleDesc, Status) VALUES ('Admin', 1);"
```
