# Citas, disponibilidad y capacidad

Estado: diseño mínimo del dominio. Reúne decisiones de la conversación y propuestas
pendientes de validación.

## 1. Objetivo

Definir las decisiones, las reglas, el esquema mínimo y API que permita
reservar una cita para un paciente, con un profesional y dentro de una clínica,
sin superponer citas ni reservar fuera del horario disponible.

## 2. Conceptos

| Concepto | Significado | Ejemplo |
| --- | --- | --- |
| Tipo de cita | Servicio reservable con una duración configurada | Consulta de 30 minutos |
| Disponibilidad | Horario durante el que el profesional puede atender en una clínica | Lunes de 09:00 a 13:00 |
| Profesional | persona especializada para otorgar la cita a pacientes. | Terapeuta o Doctor |
| Bloqueo local | Intervalo no reservable para un profesional en una clínica | Reunión en la clínica A de 11:00 a 12:00 |
| Membresias | relacion entre profesional como ususario y clinica. | Toctor A trabaja en clinica A |
| Clinica | lugar donde el profesional otorga las citas. | Clinica, Hospital o consultorio privado |
| Ususarios | usuario de la aplicacion. | Administrador, profesional o recepcionista |
| Reserva | cita creada que bloquea el tiempo de esta. | Cita de paciente A con profesional A en clinica A con tiempo de inicio y final |
| Ausencia general | Intervalo en el que el profesional no puede atender en ninguna clínica | Vacaciones toda la semana |
| Cita | Reserva con paciente, profesional e intervalo concreto | Luis con Ana, de 09:20 a 09:50 |
| Serie recurrente | Patrón semanal que origina citas individuales | Luis con Ana, martes y jueves a las 16:00 |
| Capacidad | Tiempo de atención aprovechable y posibilidad de acomodar citas de una duración | Huecos continuos donde cabe una consulta de 30 minutos |

Complementa el diseño de
[usuarios, membresías y profesionales](usuarios-membresias-profesionales.md).

## Esquema

### `appointment_types`: tipos de cita por clínica

Campos: `id`, `clinic_id`, `name`, `duration_minutes`, `active`, `created_at`,
`updated_at`.

- `duration_minutes` es un entero entre 15 y 60 inclusive. El backend y la base de
  datos deben rechazar valores fuera del rango.
- El administrador configura los tipos y su duración.
- Inicialmente, quien agenda selecciona un tipo y no modifica libremente su duración.
- Desactivar un tipo impide nuevas reservas con él; no elimina las existentes.

### `professional_appointment_types`: servicios habilitados por profesional

Campos: `clinic_id`, `professional_id`, `appointment_type_id`, `active`,
`created_at`, `updated_at`.

- La combinación `(clinic_id, professional_id, appointment_type_id)` es única.
- El tipo pertenece a esa clínica y el profesional debe estar habilitado para
  atender allí. Las referencias compuestas y validaciones deben impedir mezclar clínicas.
- Solo un administrador de la clínica gestiona estas habilitaciones.
- Tener rol profesional no habilita automáticamente todos los tipos de cita.


### `professional_availability`: disponibilidad semanal con vigencia

Campos: `id`, `clinic_id`, `professional_id`, `weekday`, `start_local_time`,
`end_local_time`, `valid_from`, `valid_until`, `active`, `created_at`, `updated_at`.

- Relaciona una clínica y un profesional habilitado para atender allí.
- `weekday` usa una convención explícita: propuesta de lunes = 1 a domingo = 7.
- Las horas se interpretan en la zona de la clínica, como `America/Mexico_City`.
- La vigencia usa fechas locales; se propone incluir ambos extremos y permitir
  `valid_until` vacío cuando no existe fecha final.
- Permite varios tramos en un día, por ejemplo 09:00–13:00 y 15:00–18:00.
- Para la primera versión se proponen tramos que terminan el mismo día y no se
  superponen en la misma clínica durante una vigencia coincidente.
- Conservar la vigencia permite cambiar horarios sin reescribir el pasado.

### `professional_blocks`: excepciones concretas

Campos: `id`, `clinic_id`, `professional_id`, `starts_at`, `ends_at`, `reason`,
`active`, `created_at`, `updated_at`.

- Bloquea un intervalo completo para la agenda indicada: reunión, ausencia, etc.
- Afecta únicamente a la clínica indicada. Las ausencias generales se representan
  por separado y también se descuentan al consultar disponibilidad.
- Crear un bloqueo no cancela automáticamente citas existentes que coincidan.
  Para cambios ordinarios se rechaza el conflicto hasta resolver las citas.
  Una urgencia permite un flujo explícito que cancela las citas futuras afectadas
  sin aprobación del paciente y registra el bloqueo, con auditoría y control de
  concurrencia. No es un efecto secundario silencioso de editar disponibilidad.
- El motivo debe ser administrativo, sin diagnósticos ni información clínica sensible.

### `professional_absences`: ausencias generales

Esquema propuesto: `id`, `professional_id`, `starts_at`, `ends_at`, `reason`,
`active`, `created_by`, `updated_by`, `created_at`, `updated_at`.

- No tiene `clinic_id`: describe la indisponibilidad de la persona, no el cierre
  de una clínica. El intervalo debe terminar después de su inicio.
- Mientras esté activa impide nuevas reservas y reprogramaciones al intervalo en
  todas las clínicas del profesional, incluido su consultorio privado. También
  aplica a la generación de ocurrencias recurrentes.
- Las consultas descuentan la unión de bloqueos locales y ausencias generales;
  un intervalo superpuesto no se resta dos veces de la capacidad.
- No cancela por sí sola citas existentes. Cada clínica identifica y resuelve
  sus citas afectadas. El flujo ordinario debe resolver conflictos antes de
  confirmar un cambio incompatible; una urgencia puede registrar la ausencia
  inmediatamente para impedir nuevas reservas y dejar las citas existentes
  señaladas para resolución explícita, sin borrarlas ni cancelarlas en silencio.
- Como política de autoría acordada para la primera versión, el profesional
  registra y gestiona únicamente sus propias ausencias. Un administrador local
  no obtiene facultad global por administrar una clínica. No se incluye delegación
  para registrar ausencias en nombre de otra persona, tampoco para recepción.
  Un administrador que además sea profesional puede gestionar las propias por
  esa identidad, no las de otros profesionales.
- Para calcular disponibilidad, una clínica conoce solo la indisponibilidad
  necesaria del profesional vinculado a ella, no sus otras clínicas, pacientes
  ni el motivo privado de la ausencia. Esto requiere respuestas y permisos limitados.
- Desactivar una ausencia libera su efecto sobre disponibilidad, pero no restaura
  citas canceladas ni sobrescribe reprogramaciones.

### `appointments`: citas

| Campo | Significado o referencia |
| --- | --- |
| `id` | Identificador de la cita; propuesta UUID |
| `clinic_id` | Clínica de la cita, referencia a `clinics` |
| `patient_id` | Paciente; conservar el tipo de ID existente |
| `professional_id` | Profesional titular, referencia a `professionals` |
| `appointment_type_id` | Tipo de cita perteneciente a la misma clínica |
| `starts_at` | Fecha y hora de inicio del intervalo reservado |
| `ends_at` | Fecha y hora final calculada por el backend |
| `status` | Estado de la cita |
| `series_id`, `series_slot_id` | Referencias opcionales a la serie y al espacio semanal de origen |
| `recurrence_date` | Fecha local original de la ocurrencia; no cambia al reprogramarla |
| `is_recurrence_exception` | Indica que la ocurrencia tiene una modificación individual respecto al patrón |
| `cancellation_category`, `cancellation_reason` | Categoría `ordinary` o `professional_emergency` y motivo administrativo cuando se cancela |
| `cancelled_by`, `cancelled_at` | Actor autenticado y momento de la cancelación |
| `patient_agreement_reference` | Referencia al evento que registra la solicitud o aprobación ordinaria del paciente, con canal, fecha y actor |
| `cancellation_notice_status` | Para una cancelación urgente: `pending` o `notified`; sin aviso confirmado no se marca `notified` |
| `created_by`, `updated_by` | Usuario autenticado que realiza la operación |
| `created_at`, `updated_at` | Marcas de tiempo de auditoría |

El intervalo persistido conserva la duración reservada: no se recalcula al leer
la cita usando la configuración actual del tipo. No es necesario guardar un
contador adicional de duración que pueda contradecir `starts_at` y `ends_at`.
El intervalo de una reserva también debe respetar el rango de 15 a 60 minutos.
No debe existir una restricción única incondicional por paciente/profesional/día:
impediría reemplazar una cita cancelada o tener varias citas válidas en un día.

Las fechas de citas, bloqueos y auditoría se proponen como `timestamptz`. La API
deberá recibir fechas con offset explícito o UTC, no horas sin zona. La
disponibilidad semanal usa horas locales, no una hora UTC fija para todo el año.
No se deben convertir fechas ambiguas o inexistentes silenciosamente al resolver
horarios locales. Cambiar la zona de una clínica no desplaza reservas ya guardadas.

Además de las claves foráneas, hay que garantizar que paciente y tipo pertenezcan
a la clínica indicada y que el profesional esté habilitado allí. Tener IDs válidos
por separado no garantiza una relación válida entre ellos.

### `appointment_series` y `appointment_series_slots`: recurrencia semanal

Esquema propuesto para el requisito de recurrencia:

- `appointment_series`: `id`, `clinic_id`, `patient_id`, `professional_id`,
  `appointment_type_id`, `timezone`, `duration_minutes`, `starts_on`, `ends_on`,
  `materialized_through`, `status`, `created_by`, `updated_by`, `created_at`, `updated_at`.
- `appointment_series_slots`: `id`, `series_id`, `weekday`, `start_local_time`.
  La combinación `(series_id, weekday, start_local_time)` es única.
- Estados propuestos de la serie: `active` y `stopped`. No sustituyen los estados
  de las citas individuales ni indican que una sesión fue realizada.
- Una serie guarda paciente, profesional, clínica, tipo, zona y duración de 15 a
  60 minutos. Sus espacios semanales pueden ser martes 16:00 y jueves 17:30.
  El profesional es el titular por defecto; una ocurrencia puede tener otro
  profesional mediante una excepción explícita. Para patrones regulares de
  distintos profesionales o tipos se crean series distintas.
- `ends_on` es opcional: nulo significa sin fecha final. Si existe, no puede ser
  anterior a `starts_on`. Una cita independiente no requiere una serie.
- La duración y zona se conservan en la serie. Cambiar la configuración del tipo
  o la zona de la clínica no altera silenciosamente las ocurrencias de una serie.
- Las fechas locales se convierten en instantes según la zona guardada para cada
  ocurrencia; no se genera la semana siguiente sumando siempre 168 horas en UTC.

Las series indefinidas se representan mediante el patrón, no mediante infinitas
citas. Las ocurrencias se guardan por períodos mensuales sucesivos. El patrón
semanal continúa entre períodos: generar un nuevo mes no exige crear otra serie
ni perder su historial. `materialized_through` registra hasta qué fecha local se
generaron o resolvieron explícitamente las ocurrencias, y solo avanza tras una
operación exitosa. Si existe `ends_on`, se respeta como límite.

La unidad mensual está acordada; falta precisar si corresponde al mes calendario
o a un período mensual contado desde el inicio de la atención, cuándo se genera
el siguiente período y si esa ampliación requiere confirmación de continuidad o
se ejecuta automáticamente mientras la serie esté activa. No se presupone renovación
automática ni una condición de pago para reservar. Si la serie comienza a mitad
del período, se conserva su fecha de inicio y nunca se generan citas pasadas.

El horizonte total de la terapia no está fijado: depende del avance del paciente.
Por eso `ends_on` puede permanecer nulo. El motivo de negocio es organizar espacios
que se cobran mensualmente; no se definen aquí tarifas, prorrateos, pagos, facturas
ni consecuencias de impago. Esas políticas no se deducen de la generación de citas.

El sistema no debe interpretar como libre una fecha solo porque aún no existen sus
filas de citas. Las búsquedas y reservas deben considerar las series activas y sus
excepciones. Para confirmar una ocurrencia fuera del período ya guardado se debe
resolver la ampliación mensual conforme a la política que se acuerde; mientras
no se haya resuelto, no se confirma ni se ofrece su espacio como libre por falta
de filas. También se deben detectar conflictos entre patrones recurrentes;
validar únicamente la primera ventana no garantiza que nunca choquen después.
Un conflicto al ampliar se registra y comunica, sin omitir la fecha silenciosamente,
sin sustituir otra reserva y sin avanzar el cursor sobre un conflicto no resuelto.
Reintentar la generación no produce citas en fechas ya pasadas ni inventa asistencias.

En `appointments`, los tres campos de recurrencia son todos nulos para una cita
independiente o todos obligatorios para una ocurrencia. La combinación
`(series_slot_id, recurrence_date)` debe ser única, incluso si la cita está
cancelada, y el espacio debe pertenecer a la serie indicada. Conservar esta clave
evita que un reintento regenere una ocurrencia cancelada o reprogramada. Las
relaciones de paciente, clínica y tipo corresponden a la serie. El profesional,
el día y la hora se toman inicialmente del patrón, pero pueden cambiar para una
sola cita mediante una excepción auditada (`is_recurrence_exception = true`).
El nuevo profesional debe estar habilitado en la misma clínica para el tipo de
cita. No se exige que el profesional real de una excepción siga siendo el titular
por defecto de la serie. La clave de origen y la fecha original se conservan.

### `appointment_events`: historial de cambios

Propuesta de registro inmutable: `id`, `appointment_id`, `actor_user_id`,
`event_type`, `previous_status`, `new_status`, `reason`, `occurred_at` y un
`operation_id` para agrupar operaciones sobre varias citas. Las reprogramaciones
guardan también los intervalos anterior y nuevo. El servidor establece la auditoría
y la guarda junto con el cambio; el cliente no elige el actor ni borra eventos.

Registrar una urgencia no requiere guardar detalles médicos privados del profesional;
se conserva únicamente el motivo administrativo necesario.

La solicitud o aprobación ordinaria puede registrarse como un evento de contacto:
recepción indica el canal (teléfono u otro canal operativo) y que el paciente la
solicitó o aprobó; el servidor guarda actor y fecha de registro. No se exige cuenta
del paciente, grabación de la llamada ni confirmación digital adicional. Si se
consigna una hora de contacto anterior al registro, se distingue de la marca de
auditoría que fija el servidor. `patient_agreement_reference` debe referir un evento
de la misma cita y clínica, no texto arbitrario ni un contacto de otro paciente.

El aviso de cancelación urgente se registra también como evento con canal, actor
y fecha cuando se realiza. Un intento sin contacto efectivo puede registrarse,
pero conserva `cancellation_notice_status = pending`. Solo el registro explícito
de aviso realizado permite cambiar a `notified`. No se considera un mensaje enviado
por el mero hecho de guardar la cancelación; no se incorpora envío automático.


## 3. Decisiones acordadas hasta ahora

- Cada cita individual relaciona un paciente, un profesional y una clínica.
- Un profesional atiende a un solo paciente a la vez.
- Un profesional puede trabajar en varias clínicas, incluido su consultorio privado.
- Cada profesional tiene obligatoriamente una cuenta de usuario.
- Un administrador no necesita ser profesional: puede ser el dueño de la clínica
  sin ser doctor ni terapeuta. Su cuenta y rol administrativo son suficientes.
- Las citas pueden comenzar en el minuto elegido por el usuario: por ejemplo,
  de 09:20 a 09:50. No se exige iniciar en horas exactas.
- La duración será configurable entre **15 minutos sin valor maximo**.
  No se ha impuesto que deba ser múltiplo de 15; se proponen minutos enteros.
- La duración se configura por tipo de cita dentro de cada clínica.
  Terapia de lenguaje puede durar 60 minutos y una consulta médica 30 minutos.
- Debe existir una relación explícita entre el profesional y los tipos de cita
  que puede atender en esa clínica.
- El backend calculará el final desde el inicio y la duración aplicable.
- Una reserva conservará su intervalo aunque después cambie la duración del tipo.
- No se requieren márgenes entre citas; dos intervalos contiguos son válidos.
- Una cita cancelada no ocupa espacio. Se puede crear otra con el mismo día,
  paciente, profesional, clínica e incluso el mismo intervalo, si sigue disponible.
- No hay un plazo de anticipación definido para cancelar o reprogramar; no se
  agregan restricciones arbitrarias de 24 o 48 horas.
- No se pueden crear citas con inicio en el pasado ni reprogramarlas hacia el pasado.
  El historial corresponde a citas previamente reservadas y cerradas con estado final.
- La recurrencia es opcional: se permiten citas individuales y series semanales
  con uno o varios espacios, por ejemplo martes y jueves a las 16:00.
- La fecha final de una serie es opcional. Sin fecha final, continúa indefinidamente
  hasta que se detenga explícitamente; no se crean infinitas filas de citas.
- Las ocurrencias recurrentes se guardan por períodos mensuales. Esto permite
  organizar el espacio que se cobra por mes, sin fijar cuántos meses durará la
  terapia: su continuidad depende del avance del paciente. El período de generación
  no es la fecha final del tratamiento ni implica implementar cobros.
- Una cita recurrente puede cambiar de profesional, día y hora. El usuario debe
  elegir si cambia solo la seleccionada o también las futuras de la recurrencia;
  no se modifica el historial ya finalizado.
- Los roles se combinan dentro de una clínica: administrador, profesional y recepción.
- Recepción puede gestionar una o varias agendas asignadas y registrar o actualizar
  datos administrativos de pacientes, sin acceso implícito al expediente clínico.
- Administrador y recepción pueden marcar citas como `completed` y `no_show`
  dentro de sus respectivos alcances de agenda; esto no concede acceso al expediente.
- Para sesiones de 60 minutos, `completed` puede registrarse manualmente desde
  los 45 minutos transcurridos y `no_show` desde los 30. Son umbrales desde el inicio
  programado, no cambios automáticos ni minutos fijos del reloj. La regla para
  duraciones menores de 60 minutos sigue pendiente.
- El rol `professional` por sí solo no puede marcar `completed` ni `no_show`.
  Si la misma persona tiene un rol administrativo autorizado, actúa con ese permiso.
- Ante una urgencia del profesional se pueden cancelar citas futuras sin aprobación
  del paciente. Esta excepción no se extiende a cancelaciones ordinarias.
- Administrador y recepción pueden ejecutar esa cancelación dentro de sus agendas
  autorizadas; recepción no necesita aprobación previa del administrador.
- Recepción puede registrar una solicitud ordinaria del paciente recibida por
  teléfono u otro canal, guardando canal, fecha y actor, sin confirmación digital obligatoria.
- Tras una cancelación urgente se registra aviso pendiente o realizado. Informar al
  paciente no implica esperar su aprobación para cancelar ni enviar mensajes automáticamente.
- Se incluyen bloqueos por clínica y ausencias generales del profesional. Una
  ausencia general impide nuevas reservas en todas sus clínicas durante el intervalo.
- Cada profesional registra y gestiona sus propias ausencias generales. En esta
  primera versión no se delega esa facultad: ser administrador local o recepción
  no permite registrar ausencias generales en nombre de otro profesional.
- Una cita vencida sin estado final queda pendiente de cierre administrativo;
  nunca se infiere automáticamente que el paciente asistió o faltó.
- Los cambios ordinarios de disponibilidad que entren en conflicto con citas
  futuras se rechazan hasta resolverlas explícitamente. Una urgencia permite el
  flujo de cancelación auditada, no una cancelación automática por editar horarios.
- Crear una cita, cambiar su tipo o profesional y crear una serie requieren una
  relación activa. Desactivarla no borra ni cancela citas ya reservadas; su revisión
  es explícita y no debe impedir cerrar o cancelar esas citas existentes.

Las secciones siguientes concretan una propuesta de funcionamiento. Los límites
todavía no acordados se enumeran al final para no convertirlos en reglas implícitas.

## 5. Reglas para crear y reprogramar

1. El usuario debe estar autenticado y autorizado para gestionar la agenda.
2. Para crear o reprogramar, la clínica, el paciente y el profesional deben estar
   habilitados. Desactivar al titular o su tipo no impide que un administrador o
   recepción autorizados cancelen o cierren citas existentes.
3. El paciente y el tipo de cita deben pertenecer a la clínica solicitada; el
   profesional debe tener su membresía activa, rol profesional y relación activa
   con ese tipo de cita en la clínica.
4. El final debe corresponder a la duración reservada, entre 15 y 60 minutos.
5. El intervalo completo debe caber en la disponibilidad vigente de esa clínica.
6. No puede coincidir con un bloqueo local ni una ausencia general aplicable.
7. El profesional no puede tener otra cita superpuesta, incluso en otra clínica.
8. El mismo registro de paciente no puede tener citas superpuestas. Si una persona
   tiene registros distintos en clínicas diferentes, detectar que es la misma
   persona requiere un diseño de identidad que está fuera del alcance actual.
9. Reprogramar requiere repetir las validaciones excluyendo del conflicto la propia
   cita. El cambio debe ser atómico: no se libera la reserva anterior si falla.
10. Si solo cambia la fecha o la hora, se propone conservar la duración ya reservada.
    Cambiar explícitamente el tipo de cita recalcula el intervalo y vuelve a validarlo.
11. Al mover una cita a otro profesional, se necesita permiso sobre ambas agendas.
    Trasladar una cita entre clínicas queda fuera de esta primera versión. En una
    cita recurrente se permite cambiar profesional, día y hora para esa ocurrencia
    o para las futuras seleccionadas, con alcance explícito. Cambiar el tipo sigue
    requiriendo una operación explícita y nueva validación de duración.
12. Cancelar conserva el registro y libera el horario; no equivale a borrarlo.
    Una nueva cita puede repetir paciente, profesional, clínica, día e intervalo
    de la cancelada si no hay otro conflicto. Se crea otro ID; no se resucita la
    cancelada ni se elimina su auditoría.
13. No se exige anticipación mínima para cancelar o reprogramar. Esto no autoriza
    modificar estados finales. Crear o reprogramar con inicio en el pasado se
    rechaza para todos los roles, comparando con la hora del servidor, no del cliente.
    La misma regla aplica a ocurrencias generadas: no se crean reservas retroactivas.

Se proponen intervalos con inicio incluido y fin excluido: `[inicio, fin)`.
Una cita de 09:00 a 10:00 permite otra a las 10:00: no se agrega margen adicional.
Una cita de 09:30 a 10:30 sí se superpone con la primera.

Dos intervalos se superponen cuando:

```text
inicio_A < fin_B y inicio_B < fin_A
```

No basta con consultar si está libre y después insertar: dos peticiones pueden
leer el mismo hueco a la vez. La implementación debe impedir el conflicto también
en la base de datos y coordinar las escrituras de citas, horarios y bloqueos.
La comprobación de disponibilidad es una ayuda; la reserva definitiva debe
volver a validar de forma segura ante concurrencia.

### Operación de las series

- Cada ocurrencia se guarda como una cita con estado propio. Una serie de martes y
  jueves genera las fechas de ambos espacios que correspondan a cada período
  mensual, con las excepciones que se registren. No se supone que un mes tenga
  exactamente cuatro semanas ni que todos los meses tengan el mismo número de citas.
- Antes de confirmar cada ventana de generación se validan todas sus ocurrencias:
  disponibilidad, bloqueos, habilitación de tipo, conflictos de paciente y profesional,
  incluyendo conflictos entre los espacios de la propia serie.
- Se propone creación atómica de la serie y su ventana inicial: si una ocurrencia
  de esa ventana tiene conflicto, no se confirma parcialmente. Las ampliaciones
  posteriores son atómicas por ventana; fallar no invalida las reservas anteriores.
  No se omiten fechas silenciosamente. Se limita el tamaño de cada ventana, no la
  duración total de la recurrencia, y se controla concurrencia entre generadores
  y reservas independientes.
- Reintentar la misma creación no debe duplicar serie ni citas. Se propone una
  clave de idempotencia con alcance de clínica y actor, además de las claves de ocurrencia.
- Cancelar una ocurrencia libera solo ese intervalo y conserva las otras semanas.
  Reprogramar una ocurrencia puede cambiar profesional, día y hora, pero conserva
  su clave de origen y no cambia el patrón semanal. El generador no debe restaurar
  su profesional o intervalo originales ni duplicarla.
- Una reserva de reemplazo para una ocurrencia cancelada puede ser una cita
  independiente, sin reutilizar la clave de ocurrencia, opcionalmente vinculada a
  la cancelada para trazabilidad. Esa excepción no modifica la serie.
- La interfaz y la API distinguen explícitamente «esta cita» de «esta y las futuras».
  No se deduce una modificación masiva de la edición de una sola cita.
- «Esta cita»: actualiza solo la ocurrencia elegida, incluso si se asigna a otro
  profesional, conservando su ID, su vínculo con la serie y el historial del cambio.
- «Esta y las futuras»: aplica el nuevo profesional/día/hora a las ocurrencias
  futuras del patrón seleccionado, incluyendo las todavía no generadas. Conserva
  las citas pasadas, finalizadas o canceladas y no reactiva excepciones anuladas.
  Las excepciones futuras ya reprogramadas se muestran para decisión explícita;
  no se sobrescriben sin indicarlo.
- Si la serie tiene dos espacios semanales, se propone seleccionar qué espacios
  se modifican. Cambiar el martes no mueve el jueves por accidente ni fusiona dos
  citas en una. Antes de confirmar se muestran alcance y conflictos de la operación.
- Para detener una serie se propone indicar una fecha efectiva, cancelar de forma
  explícita las ocurrencias `scheduled` afectadas y conservar las anteriores o
  finalizadas. No basta con cambiar `status` en la serie dejando reservas ocultas.
- Para cambiar el patrón futuro se propone cerrar la serie anterior y crear una
  nueva versión/serie desde la fecha efectiva, manteniendo el vínculo de origen
  y las excepciones. Los espacios semanales no seleccionados continúan intactos.
  Se actualizan patrón y ventana materializada en una operación atómica y auditada;
  si las nuevas citas no son válidas, se mantienen las reservas anteriores. Tanto
  la versión anterior como las nuevas siguen considerando reservas aún no materializadas.

### Cancelación por urgencia del profesional

- Solo una urgencia permite cancelar una cita futura sin solicitud o aprobación
  del paciente. Fuera de esa excepción recepción puede registrar la solicitud o
  aprobación recibida por teléfono u otro canal, con fecha y actor. No se requiere
  aprobación digital del paciente ni una cuenta suya en el sistema.
- La cancelación es una acción explícita sobre citas `scheduled`, con motivo,
  categoría, actor y fecha. No cambia a `no_show`: el paciente no es responsable
  de una ausencia del profesional.
- Administrador y recepción pueden ejecutar esta excepción dentro de su alcance
  habitual. Recepción no necesita autorización previa del administrador para
  cancelar. El rol profesional por sí solo no recibe esta facultad en el diseño
  actual; registrar una ausencia propia es una operación distinta de cancelar
  sin aprobación del paciente. La urgencia no amplía el acceso a otras agendas.
- Si la ausencia afecta varias fechas, se seleccionan las ocurrencias concretas.
  No se cancela toda la serie por defecto ni se cambian sesiones ya realizadas.
- La cancelación libera la reserva, pero un bloqueo de ausencia sigue haciendo
  que el profesional no esté disponible en ese intervalo. No se debe ofrecer ese
  hueco mientras exista el bloqueo; sigue libre para reservar solo si cumple las
  demás reglas. Una operación conjunta de cancelación y bloqueo local debe
  persistirse atómicamente. Una ausencia general urgente puede impedir nuevas
  reservas de inmediato, mientras cada clínica resuelve sus citas existentes;
  no se condiciona ese bloqueo global a poder cancelar citas en otras clínicas.
- Si la operación incluye crear un bloqueo, el actor debe tener permiso para
  bloquear esa agenda. Recepción puede cancelar por sí sola, pero no recibe permiso
  general para crear bloqueos o ausencias de otra persona. La operación conjunta
  requiere un actor autorizado para ambas acciones; no es requisito que un
  administrador apruebe cada cancelación realizada por recepción.
- Avisar al paciente es distinto de pedirle aprobación: se registra
  comunicación pendiente/realizada mediante el canal operativo de la clínica,
  sin bloquear la cancelación urgente por falta de respuesta. No se implementa
  aquí un sistema automático de mensajes ni se afirma que el aviso fue enviado.

## 6. Estados propuestos

| Estado | Significado | Transición inicial propuesta |
| --- | --- | --- |
| `scheduled` | Reserva programada | Puede reprogramarse o pasar a uno de los estados finales |
| `completed` | Sesión realizada | Final; no volver a programada por un PATCH genérico |
| `cancelled` | Reserva cancelada | Final; libera el horario sin borrar el registro |
| `no_show` | Paciente no asistió | Final; conserva el intervalo que se había reservado |

Una cita nueva inicia en `scheduled`; no se acepta un estado arbitrario enviado
por el cliente. Las transiciones son operaciones con reglas, no texto libre.

Para conservar la coherencia del historial se propone que solo `cancelled` deje
de contar como intervalo ocupado al verificar conflictos. `completed` y `no_show`
no equivalen a huecos históricos que puedan reutilizarse.

Administrador y recepción pueden marcar `scheduled` como `completed` o `no_show`
dentro de su alcance de agenda. Marcar asistencia es una operación administrativa,
no autoriza leer ni escribir notas clínicas. El rol profesional por sí solo no
puede realizar ninguno de esos cierres. Una persona con varios roles sí puede
hacerlo por un rol administrativo autorizado en la clínica correspondiente.
Los cierres siempre requieren una acción manual del usuario autorizado; el paso
del tiempo solo habilita la acción, no demuestra el resultado de la sesión.
Para una cita de **60 minutos**, se acuerdan estos límites, ambos inclusive:

- `completed`: desde `starts_at + 45 minutos`, si el usuario registra asistencia.
- `no_show`: desde `starts_at + 30 minutos`, si el usuario registra inasistencia.

Por ejemplo, para una sesión de 09:20 a 10:20, `no_show` se habilita a las 09:50
y `completed` a las 10:05. No se interpretan como las 09:30 o las 09:45.
El servidor comprueba el tiempo transcurrido con su propio reloj y guarda actor,
fecha y transición. Llegar al umbral no autoriza inferir asistencia o inasistencia.
Marcar el resultado antes del final no acorta `ends_at` ni libera el resto del
intervalo reservado; `completed` registra aquí la asistencia administrativa, no
una medición de la hora real de salida ni de los minutos efectivamente atendidos.

Para tipos de **15 a 59 minutos**, falta acordar los umbrales. No se aplican
porcentajes, límites al final de la cita ni los mismos 45/30 minutos por defecto.
Esa decisión debe cerrarse antes de habilitar los cierres de esos tipos.
Las correcciones de estados finales requieren un flujo explícito con auditoría
y quedan fuera del CRUD genérico.

No se crean citas retroactivamente, ni siquiera enviando un estado final. El
historial se conserva a partir de reservas que existían antes y que posteriormente
se cerraron como `completed`, `cancelled` o `no_show`.
El objetivo operativo es que las citas pasadas queden en estado final. El paso del
tiempo no demuestra asistencia o inasistencia: si una cita vence aún en `scheduled`,
se identifica como pendiente de cierre administrativo, sin borrarla ni cambiarla
automáticamente a `no_show`. Debe resolverse antes de considerarla historial cerrado.
Este pendiente es una condición derivada (`status = scheduled` y `ends_at` anterior
o igual a la hora actual), no un quinto estado de asistencia. Se muestra en una
bandeja filtrada por las agendas autorizadas de administrador y recepción, quienes
registran el resultado real mediante las transiciones ya permitidas. La auditoría
guarda cuándo se realizó el cierre, aunque haya sido después de la sesión.

## 7. Capacidad y horarios disponibles

Con duraciones variables, capacidad no significa un número fijo de citas por día.
Se deben distinguir:

- Tiempo habilitado para atender, según disponibilidad y bloqueos.
- Tiempo reservado por citas no canceladas.
- Intervalos libres continuos donde cabe el tipo de cita solicitado.
- Tiempo efectivamente atendido, a partir de citas realizadas.

Para buscar horarios se parte de la disponibilidad vigente, se descuentan los
bloqueos aplicables y los intervalos ocupados del profesional, y se buscan huecos
continuos de la duración solicitada. Si ya se seleccionó paciente, también se
descartan sus conflictos. Una agenda libre en A no ignora una cita del profesional
en B. Los detalles de esa otra clínica no se muestran a quien no tiene permiso.

Ejemplo: un profesional atiende de 09:00 a 13:00, tiene un bloqueo de 11:00 a 12:00
y una cita de 09:20 a 09:50. Quedan 150 minutos libres distribuidos en tres tramos:
09:00–09:20, 09:50–11:00 y 12:00–13:00. Una cita de 30 minutos no cabe en el primero.

Dos huecos separados de 15 minutos no permiten una sesión continua de 30 minutos.
Tampoco se deben sumar horarios superpuestos en distintas clínicas como si fueran
capacidad independiente de una misma persona. El cómputo agregado entre clínicas
requiere descontar esa duplicación. El diseño actual no agrega tiempos de traslado.

Por ahora la capacidad se deriva de horarios, bloqueos y citas; no se guarda un
contador mutable de espacios libres. Los reportes avanzados quedan para después.

## 8. Autorización e integración con los otros módulos

- Administrador: gestiona las agendas de su clínica, no las de otras clínicas.
  No necesita perfil profesional. Administra las habilitaciones profesional–tipo.
- Profesional: gestiona su propia agenda en las clínicas donde está habilitado.
  Su rol por sí solo no permite marcar `completed` o `no_show`.
- Recepción: gestiona únicamente las agendas asignadas dentro de su clínica.
- Administrador y recepción pueden marcar `completed` y `no_show` y gestionar
  recurrencias dentro de ese mismo alcance. Una serie no amplía sus permisos.
- La cancelación urgente sin aprobación requiere la validación y auditoría
  específicas anteriores; no equivale a autorización para cancelar indiscriminadamente.
- Recepción no necesita aprobación del administrador para esa cancelación en una
  agenda asignada. Puede registrar la solicitud ordinaria del paciente y el aviso
  pendiente/realizado, sin acceder a expedientes clínicos.
- Un bloqueo local no afecta otras clínicas. Una ausencia general limita reservas
  en todas ellas, pero no permite a sus administradores ver o modificar citas ajenas.
- Recepción puede registrar y editar datos administrativos de pacientes; eso no
  concede acceso a citas en agendas no asignadas ni al expediente clínico.
- El backend obtiene actor y permisos de la identidad autenticada; no confía en
  roles, propietario o campos de auditoría enviados como prueba por el cliente.
- Tener `clinic_id` en una ruta no sustituye la verificación de acceso.

Las reglas puras de intervalos se pueden implementar y probar sin conexión a
Supabase. Antes de exponer endpoints con datos reales, deben existir autenticación,
autorización y protecciones de persistencia, incluyendo los módulos ya existentes.

## 9. Casos de aceptación para la futura implementación

- Aceptar una cita de 30 minutos que inicia a las 09:20 y termina a las 09:50.
- Aceptar duraciones de 15 y 60 minutos y rechazar 14, 61, cero o negativas,
  tanto en tipos como en reservas y series.
- Rechazar una cita cuyo final rebasa el horario de atención.
- Rechazar altas o reprogramaciones hacia el pasado, incluso con rol `admin` o
  con un estado final enviado por el cliente. El generador tampoco crea citas pasadas.
- Rechazar una cita que atraviesa un descanso o coincide con un bloqueo.
- Rechazar un traslape del profesional aunque la otra cita sea en otra clínica.
- Rechazar un traslape del mismo registro de paciente.
- Permitir dos citas contiguas sin agregar márgenes.
- Rechazar paciente o tipo de cita perteneciente a una clínica distinta.
- Rechazar un tipo que no esté habilitado para el profesional en esa clínica.
- Conservar las reservas existentes al cambiar la duración configurada del tipo.
- Conservar la reserva original si falla una reprogramación.
- Cancelar y poder reservar después el intervalo liberado, conservando el historial.
- Repetir exactamente día, paciente, profesional, clínica e intervalo de una
  cancelada en una nueva cita con otro ID, si no hay bloqueo ni otro conflicto.
- Permitir cancelar o reprogramar cerca del inicio sin imponer un plazo inventado.
- Rechazar una transición de estado no permitida.
- Permitir que solo una de dos reservas simultáneas conflictivas se guarde.
- Evitar que cambios concurrentes de horarios o bloqueos dejen reservas inválidas.
- Rechazar acceso de recepción a una agenda no asignada y a una clínica ajena.
- No ofrecer una sesión de 30 minutos usando dos huecos separados de 15 minutos.
- Permitir a un administrador sin perfil profesional gestionar la agenda y tipos.
- Permitir `completed` y `no_show` a administrador y recepción autorizados y
  rechazar estas acciones en una clínica ajena o agenda no asignada.
- Rechazar esos cierres al rol profesional por sí solo, pero permitirlos por un
  rol administrativo adicional válido. No inferir `no_show` por el paso del tiempo.
- En sesiones de 60 minutos, rechazar `completed` antes de los 45 minutos desde
  el inicio y `no_show` antes de los 30; aceptar en el umbral exacto o después
  únicamente si el estado sigue siendo `scheduled` y el actor está autorizado.
- Para una sesión de 09:20 a 10:20, habilitar manualmente `no_show` a las 09:50
  y `completed` a las 10:05, usando la hora del servidor y conservando el intervalo
  completo. Sin acción del usuario permanece `scheduled` aunque se cumpla el umbral.
- Crear una cita individual sin serie y una serie sin fecha final; no obligar a
  activar recurrencia ni a elegir una fecha de término.
- Crear una serie de martes y jueves y obtener dos ocurrencias por cada semana
  completa de la vigencia, respetando la hora local y duración guardadas.
- Generar las fechas reales de cada período mensual, incluidos los meses con
  cinco ocurrencias de un mismo día semanal; no asumir cuatro semanas por mes.
- Continuar una serie sin fecha final de un período mensual al siguiente sin
  duplicarla, perder excepciones ni crear citas anteriores a su inicio o al presente.
  La política de delimitación y ampliación mensual debe acordarse antes de completar
  sus pruebas de aceptación; generar el período no acredita ningún pago.
- Rechazar atómicamente la ventana inicial si una ocurrencia tiene conflicto y,
  al fallar una ampliación, conservar las citas ya confirmadas y señalar el conflicto.
- Ampliar una serie sin fecha final sin duplicar ocurrencias ni perder excepciones,
  respetando el límite de generación y sin ofrecer fechas no generadas como libres.
- Reintentar una creación sin duplicar la serie ni sus ocurrencias.
- Cancelar o reprogramar una ocurrencia sin alterar las siguientes ni regenerar
  la cancelada al reintentar el proceso.
- Detener o cambiar la parte futura de una serie sin modificar citas finalizadas.
- Cambiar profesional, día y hora solo en una ocurrencia sin cambiar el resto.
- Aplicar esos cambios a las futuras seleccionadas, incluidas las no generadas,
  sin alterar el pasado, reactivar cancelaciones ni sobrescribir excepciones sin aviso.
- Rechazar un cambio de profesional sin habilitación para el tipo o sin permisos
  sobre las agendas de origen y destino; preservar las reservas si el cambio falla.
- Rechazar un cambio ordinario de disponibilidad que deje citas futuras fuera
  de horario hasta resolverlas explícitamente.
- Cancelar citas futuras por urgencia sin aprobación del paciente, con actor y
  motivo registrados; rechazar el uso de la excepción en una cancelación ordinaria.
- No convertir una urgencia del profesional en `no_show` del paciente.
- No ofrecer como disponible una cita cancelada cuyo intervalo sigue bloqueado.
- Un bloqueo de la clínica A no bloquea B; una ausencia general del profesional
  impide nuevas reservas en ambas, incluidas las recurrentes.
- No duplicar el tiempo descontado al superponerse un bloqueo local y una ausencia.
- Registrar una ausencia no cancela automáticamente citas ni expone información
  de otra clínica; su revocación no reactiva citas canceladas.
- Permitir que el profesional gestione su propia ausencia general y rechazar
  que un administrador local o recepción gestionen la de otro profesional.
  Combinar roles no concede esa delegación ni acceso a agendas de otras clínicas.
- Permitir a recepción cancelar por urgencia en sus agendas sin aprobación previa
  del administrador; rechazar esa operación en agendas no asignadas.
- Aceptar solicitud ordinaria por teléfono registrada con canal, actor y fecha;
  no exigir confirmación digital ni aceptar un evento de contacto de otra cita.
- Mantener aviso pendiente hasta registrar el contacto realizado; no considerar
  avisado al paciente por una cancelación o un intento fallido de contacto.
- Mostrar una cita vencida `scheduled` como pendiente de cierre sin convertirla
  automáticamente en `completed` o `no_show` y cerrarla solo con un actor autorizado.

Son criterios de prueba, no pruebas ya escritas ni resultados verificados.

## 10. Decisiones pendientes y límites de alcance

Las decisiones confirmadas se incorporan a las secciones anteriores. Esta lista
conserva únicamente lo que todavía necesita una decisión; no repite como pendiente
la generación mensual, el cierre manual de sesiones de 60 minutos ni la autoría
propia de ausencias generales.

### Detalles técnicos y operativos por cerrar

1. **Delimitar y ampliar los períodos mensuales.** Ya se decidió guardar por mes,
   no fijar una duración total de terapia. Falta elegir mes calendario (por ejemplo,
   octubre) o período mensual desde la fecha de inicio, y cuándo y quién confirma
   la generación del siguiente período. También falta concretar el mecanismo de
   reintento y aviso si la ampliación falla. Ya está definido que no se omiten
   conflictos, no se pierden reservas anteriores ni se ofrecen espacios falsamente
   libres. El pago no es una condición de generación definida en este diseño.
2. **Umbrales de cierre para citas menores de una hora.** Los 45 minutos para
   `completed` y los 30 para `no_show`, con acción manual, quedan definidos para
   sesiones de 60 minutos. Falta decidir los límites para tipos de 15 a 59 minutos:
   por ejemplo, aplicar 45 minutos a una consulta de 15 obligaría a esperar media
   hora después de su final. No se elige automáticamente una regla proporcional.
3. **Horas locales ambiguas o inexistentes (parte horaria del antiguo punto 4).**
   En una zona que adelanta el reloj, una hora local puede no ocurrir; si lo atrasa,
   una misma hora puede ocurrir dos veces. Ejemplo hipotético: si se salta de 02:00
   a 03:00, no existe la cita de las 02:30; si se repite la franja de 01:00 a 02:00,
   las 01:30 identifican dos instantes distintos. No se afirma que esos cambios
   ocurran en la zona actual de la clínica. Falta elegir cómo resolver esas fechas
   de una recurrencia. Propuesta pendiente de aprobación: señalar el conflicto
   sin mover ni omitir la cita automáticamente; si la hora se repite, exigir elegir
   uno de los dos instantes válidos. Las reservas individuales mantienen el requisito
   de fecha con offset explícito o UTC.
4. **Alcance al editar varios espacios y sus excepciones (parte de recurrencia
   del antiguo punto 4).** Si un paciente va martes y jueves, cambiar el martes
   desde cierta fecha no debe mover el jueves por accidente. Falta confirmar la
   propuesta de seleccionar «solo este espacio semanal» o «espacios seleccionados»
   dentro de «esta y las futuras», mostrando las fechas afectadas antes de guardar.
   También falta confirmar qué hacer con una cita futura que ya se movió de forma
   individual: propuesta, conservarla por defecto e incluirla solo mediante una
   selección explícita con nueva validación de conflictos. Cambiar una sola cita
   o las futuras ya está aprobado; las pasadas/finalizadas no se modifican y las
   canceladas no se reactivan. No se amplían los permisos de origen y destino.

Dentro de esta primera versión: citas individuales y recurrencia semanal opcional,
con uno o varios espacios, fecha final opcional, y excepciones o cambios en serie.

Fuera de esta primera versión: sesiones grupales, pagos, recordatorios automáticos,
listas de espera, asignación de consultorios y reportes avanzados. No se agregan
márgenes automáticos ni tiempos de traslado entre citas.

## 11. Orden de trabajo

La secuencia, las dependencias, los estados y la planificación se mantienen en el
[plan único de Notion](https://app.notion.com/p/46deaaf04e474ebab434087ae36be978).
Consultar las vistas **Secuencia** y **Para planificar** para decidir qué sigue.
Este documento conserva las reglas y los criterios técnicos; no duplica el backlog
ni acredita avances de implementación. Los detalles aún abiertos de la sección 10
se gestionan en la tarea DIS-01 del tablero.

No es necesario construir primero toda una interfaz de administración de usuarios
para avanzar en el dominio. Sí es necesario proteger el servicio antes de usarlo
con información real.
