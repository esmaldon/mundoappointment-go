# Review del backend — 5 de octubre de 2026

Se revisaron los cambios locales de clínicas, pacientes, servicios, rutas, modelos y acceso a Supabase. Se conservaron los cambios previos de README y diseño.

## Hallazgos pendientes

### P1 — El identificador del paciente no coincide con Supabase

`backend/internal/patients/model.go:6` declara `Patient.Id` como UUID. La consulta del esquema OpenAPI de la base de pruebas confirmó que `patients.id` sigue siendo `bigint`. Las respuestas con identificadores numéricos no pueden deserializarse en el modelo actual; los filtros con UUID tampoco son compatibles con esa columna.

Antes de desplegar, migrar IDs y todas sus referencias a UUID, con una correspondencia que preserve los registros existentes, o conservar el contrato numérico. No se revirtió el cambio de UUID ni se transformaron datos remotos. La migración de IDs no está incluida: requiere revisar claves foráneas y consumidores del identificador.

### P1 — La base exige tutor incluso a pacientes adultos

`backend/internal/patients/service.go:66` deja `parent_id` nulo para adultos. La integración real falló con SQLSTATE `23502`: la columna tiene una restricción `NOT NULL`. Usar un UUID de ceros no representa la ausencia de tutor y puede violar la clave foránea.

Se preparó `backend/migrations/20261005_patients_optional_parent.sql` para permitir valores nulos, conservando la clave foránea. No se ejecutó sobre Supabase. Esta migración solo resuelve la obligatoriedad del tutor, no la incompatibilidad del ID del paciente.

## Fallos corregidos

- Pruebas incompatibles con las nuevas interfaces de servicio, nombres de inicializadores, UUID y respuestas de creación de un objeto en lugar de una lista.
- Ruta de edición del tutor sin manejador; discrepancia `parentId`/`parentid`; traducción incorrecta del error de tutor inexistente.
- Edición de tutores sin comprobar que pertenecen al paciente y clínica de la URL.
- Desreferencias nulas al recibir datos incompletos del tutor y asignación de la fecha de nacimiento del paciente al tutor.
- Cálculo de mayoría de edad basado solo en el año. Ahora considera el cumpleaños y rechaza fechas futuras de nacimiento del paciente.
- Errores de validación del alta que respondían 500 en lugar de 400.
- Nombres de columnas de tutores incompatibles con la base: se comprobó que usa `first_name` y `last_name`. El acceso a datos traduce esos nombres sin cambiar el JSON público de `Parent`.
- Acceso a `result[0]` sin comprobar que la inserción devolviera exactamente un registro.
- Se intenta eliminar el tutor recién creado cuando falla la creación del paciente, conservando también el error de limpieza si falla. Esto no equivale a una transacción: interrupciones del proceso o fallos ambiguos de red requieren una operación atómica en la base para garantizar consistencia.

## Pruebas y resultados

- Compilación del backend: correcta.
- `go vet` de los cuatro módulos: correcto.
- Pruebas locales con `-race -count=1`: correctas.
- Compilación de pruebas con etiqueta `integration`: correcta.
- Integración real de clínicas (`TestStoreClinicCRUD`): correcta.
- Integración real de tutores (`TestStoreParentCRUD`): correcta después de corregir el mapeo de columnas.
- Integración real de pacientes adultos: falló por `patients.parent_id NOT NULL`.
- Primera integración de menores: detectó la incompatibilidad de columnas de tutores, ya corregida y comprobada por separado. El flujo completo de pacientes queda pendiente de la migración de IDs; no se declara aprobado.

Las pruebas de pacientes ya no dependen de una clínica fija: crean y limpian sus propias clínicas. Se añadieron casos para cumpleaños, datos incompletos del tutor, errores de almacenamiento, separación entre clínicas, asociación paciente-tutor y respuestas inválidas de inserción. Las pruebas con errores de esquema no se omitieron ni se alteraron para fingir un resultado satisfactorio.

## Repetir las verificaciones

Desde `backend`:

```sh
go build -o /tmp/mundoappointment-review-backend ./cmd
go vet ./cmd/... ./internal/patients/... ./internal/clinics/... ./pkg/...
go test -race -count=1 ./cmd/... ./internal/patients/... ./internal/clinics/... ./pkg/...
go test -tags=integration -run '^$' ./internal/patients/... ./internal/clinics/...
```

Después de alinear el esquema, ejecutar la suite con `-tags=integration -count=1 -timeout=2m`, `RUN_SUPABASE_INTEGRATION=true` y las variables `SUPABASE_TEST_URL`/`SUPABASE_TEST_KEY` del entorno de pruebas. No utilizar una base de producción.
