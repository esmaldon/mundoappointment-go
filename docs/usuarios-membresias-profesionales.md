# Usuarios, membresías y profesionales

Estado: diseño de referencia para la primera versión. El usuario reporta haber
aplicado migraciones de este dominio en Supabase; su estructura remota no se ha
verificado aquí. Las actualizaciones de este documento no aplican migraciones ni
demuestran que la autorización esté implementada.

## Objetivo y alcance

Una persona puede atender y administrar en una clínica, pero solamente atender
en otra. Los permisos deben depender de la clínica de cada petición.

Este documento define relaciones y permisos de usuarios. Se complementa con
[citas, disponibilidad y capacidad](cita-disponibilidad.md), que desarrolla
reservas, tipos de cita, recurrencia y cancelaciones. No implementa endpoints.

Decisiones ya acordadas:

- Un profesional atiende a un paciente a la vez.
- Puede trabajar en varias clínicas, incluido su consultorio privado.
- Los roles de profesional y administrador se pueden combinar por clínica.
- Cada profesional tendrá obligatoriamente una cuenta de usuario.
- Un administrador no necesita ser profesional. El dueño de una clínica puede
  tener cuenta y rol `admin` sin perfil de doctor ni terapeuta.
- Recepción necesita gestionar una o varias agendas sin obtener todos los
  permisos de un administrador.
- Recepción puede registrar pacientes nuevos y actualizar sus datos administrativos.
- Las citas pueden iniciar en cualquier minuto, sin una cuadrícula obligatoria.
- La duración se configurará por tipo de cita y se conservará en cada reserva.
- La duración mínima es de 15 minutos sin valor maximo.
- No se necesitan márgenes entre citas ni hay un plazo de anticipación definido
  para cancelar o reprogramar.
- Una cita cancelada libera su intervalo y permite otra con el mismo día,
  paciente, profesional y clínica, si no hay otros conflictos o bloqueos.
- Cada profesional tiene una relación explícita con los tipos de cita que puede
  atender en cada clínica; la membresía por sí sola no habilita todos los tipos.
- Administrador y recepción pueden marcar `completed` y `no_show` en las agendas
  que tienen autorizadas, sin obtener acceso a expedientes clínicos.
- El rol profesional por sí solo no puede marcar `completed` ni `no_show`. Si
  tiene además un rol administrativo autorizado, puede hacerlo por ese otro rol.
- No se pueden crear ni reprogramar citas hacia el pasado. El historial se conserva
  mediante citas previamente reservadas y cerradas, no mediante altas retroactivas.
- La recurrencia semanal es opcional, de uno o varios espacios por paciente, con
  fecha final opcional; puede continuar indefinidamente hasta detenerla.
- Se puede cambiar profesional, día y hora de una cita recurrente, eligiendo solo
  la cita seleccionada o también las futuras del patrón elegido, sin alterar el pasado.
- Los cambios ordinarios de disponibilidad con conflictos se rechazan hasta
  resolver las citas futuras; no se cancelan automáticamente por editar horarios.
- Una urgencia del profesional permite cancelar citas futuras sin aprobación del
  paciente. La excepción requiere motivo, actor y fecha y no se aplica a casos ordinarios.
- Administrador y recepción pueden ejecutar la cancelación urgente en sus agendas
  autorizadas. Recepción no necesita aprobación previa del administrador para ello.
- Recepción puede registrar una solicitud o aprobación ordinaria recibida por
  teléfono u otro canal, con canal, fecha y actor, sin confirmación digital obligatoria.
- El aviso de cancelación urgente se registra como pendiente o realizado. No se
  requiere respuesta del paciente para cancelar ni se presupone envío automático.
- Hay bloqueos por clínica y ausencias generales del profesional. Una ausencia
  global impide nuevas reservas en todas sus clínicas sin ampliar permisos sobre
  pacientes o citas de otras clínicas.
- Cada profesional registra y gestiona solo sus propias ausencias generales.
  No se incluye delegación a administrador o recepción en esta primera versión.
- Una cita vencida sin resultado queda pendiente de cierre administrativo, sin
  asignar automáticamente asistencia o inasistencia.
- La agenda no depende de asignar un consultorio físico.

Propuestas para mantener acotada esta primera versión:

- Usar Supabase Auth para las cuentas, sin implementar contraseñas propias.
- El administrador podrá invitar a los profesionales a crear su cuenta. No se
  habilita atención hasta tener cuenta y perfil profesional vinculados.
- Recepción tendrá su propia cuenta y membresía con rol `receptionist`, sin
  requerir un perfil profesional. Su acceso será a agendas asignadas explícitamente.
- Doctor y terapeuta serán perfiles profesionales, no roles de autorización.
- Usar desactivación, no borrado físico, para relaciones con historial.

## Modelo de datos propuesto

Los nombres siguientes son nuevos; la tabla existente de clínicas es `clinics`.
Los identificadores de estas nuevas tablas serán UUID. No se propone cambiar
el identificador actual de pacientes.

### 1. Identidad: `auth.users` y `user_profiles`

Supabase Auth controla el inicio de sesión, la contraseña y las sesiones.
`user_profiles` guarda solamente información propia de la aplicación:

| Campo | Regla |
| --- | --- |
| `id` | Clave primaria y referencia a `auth.users.id`; no generar otro UUID |
| `display_name` | Obligatorio y no vacío después de quitar espacios |
| `active` | Booleano obligatorio |
| `created_at`, `updated_at` | Fecha y hora con zona (`timestamptz`) |

El correo de acceso pertenece a Auth. No duplicar contraseñas, tokens ni roles
en este perfil. El usuario no puede modificar su estado activo libremente.

### 2. Perfil de atención: `professionals`

| Campo | Regla |
| --- | --- |
| `id` | Clave primaria UUID |
| `user_id` | Referencia obligatoria y única a `user_profiles.id` |
| `kind` | Uno de `doctor` o `therapist` para la primera versión |
| `active` | Booleano obligatorio |
| `created_at`, `updated_at` | `timestamptz` |

Un usuario puede no ser profesional (por ejemplo, un administrador o recepcionista).
La relación usuario–profesional es opcional desde el usuario y obligatoria desde
el profesional: no se crea un perfil profesional ficticio para dar permisos de
administración. El dueño no necesita tener una agenda de atención para administrarlas.
Un perfil profesional no concede permisos por sí solo. No hay un `clinic_id`
único aquí porque la persona puede trabajar en varias clínicas.

El perfil es global: un administrador de una clínica no puede desactivar a la
persona en todas las otras clínicas. Solo puede administrar su vinculación local.
La administración del perfil global requiere una política aparte, fuera del
CRUD administrativo de una clínica.

### 3. Vinculación: `clinic_memberships`

| Campo | Regla |
| --- | --- |
| `id` | Clave primaria UUID |
| `clinic_id` | Referencia obligatoria a `clinics.id` |
| `user_id` | Referencia obligatoria a `user_profiles.id` |
| `active` | Booleano obligatorio |
| `created_at`, `updated_at` | `timestamptz` |

Debe existir una restricción única sobre `(clinic_id, user_id)`. Al reincorporar
a alguien se reactiva su membresía; no se crea una segunda para la misma pareja.

### 4. Permisos asignados: `clinic_membership_roles`

| Campo | Regla |
| --- | --- |
| `membership_id` | Referencia obligatoria a `clinic_memberships.id` |
| `role` | Uno de `admin`, `professional` o `receptionist` |

La clave primaria será `(membership_id, role)`: permite combinar roles sin repetirlos.
Una membresía sin roles no concede acceso. No se necesita un catálogo de roles
editable por usuarios en esta primera versión.

Para asignar `professional`, el usuario debe tener un perfil profesional activo.
El servicio deberá comprobarlo y guardar los cambios relacionados de forma
atómica. La implementación de persistencia deberá proteger esa regla también
frente a modificaciones concurrentes o accesos directos no autorizados.

La membresía con rol `professional` es, en esta versión, la habilitación para
atender en esa clínica; no se agrega otra tabla que duplique esa pertenencia.
Además, `professional_appointment_types`, definida en el documento de agenda,
determina qué servicios puede atender allí. Son restricciones distintas: pertenecer
a una clínica no autoriza automáticamente a ofrecer todos sus tipos de cita.

### 5. Alcance de recepción: `clinic_agenda_assignments`

El rol indica qué acciones están permitidas; una asignación indica sobre qué
agenda se pueden realizar. Tener `receptionist` no concede acceso a todas las
agendas de la clínica. Sin asignaciones activas, no concede acceso a ninguna.

| Campo | Regla |
| --- | --- |
| `clinic_id` | Clínica a la que pertenecen las dos membresías |
| `receptionist_membership_id` | Membresía de quien gestionará la agenda |
| `professional_membership_id` | Membresía del profesional titular de la agenda |
| `active` | Booleano obligatorio |
| `created_at`, `updated_at` | `timestamptz` |

La pareja de membresías será única. Una recepcionista puede tener varias
asignaciones y varias recepcionistas pueden gestionar la agenda de un profesional.
No se crea una agenda duplicada: todas operan sobre las mismas citas del titular.

Ambas membresías deben pertenecer a `clinic_id`. La futura migración debe usar
referencias compuestas `(membership_id, clinic_id)` hacia una restricción única
`(id, clinic_id)` en `clinic_memberships`, para impedir asignaciones entre clínicas.
Además, al asignar se deben verificar los roles correspondientes y los estados
activos de las membresías, cuentas y profesional.

Solo un administrador de esa clínica puede crear, reactivar o revocar estas
asignaciones. No se confía en una lista de profesionales enviada por recepción
como prueba de autorización.

Esta tabla define alcance de autorización, no horarios ni disponibilidad.

## Ejemplo

Luis tiene una cuenta y un perfil profesional, pero tres membresías:

| Clínica | Roles de Luis | Alcance |
| --- | --- | --- |
| Clínica A | `admin`, `professional` | Puede atender y administrar esta clínica |
| Clínica B | `professional` | Solo su agenda y pacientes autorizados allí |
| Consultorio privado | `admin`, `professional` | Puede atender y administrar esta unidad |

Al desactivar la membresía de Clínica B, pierde ese acceso sin perder el de A
ni el de su consultorio. Los pacientes siguen separados por clínica.

María tiene cuenta y membresía en A con rol `receptionist`, pero no perfil
profesional. El administrador le asigna las agendas de Luis y Ana. Puede gestionar
esas agendas en A, pero no las de otros profesionales ni la agenda de Luis en B.
No puede asignarse agendas adicionales ni conceder roles.

Elena es dueña de A, pero no es doctora ni terapeuta. Tiene cuenta, membresía y rol
`admin`, sin fila en `professionals`. Puede administrar A y registrar asistencia
sin que el sistema le exija una especialidad o una agenda propia.

## Matriz inicial de autorización

| Acción dentro de una clínica | `admin` | `professional` | `receptionist` |
| --- | --- | --- | --- |
| Consultar agendas | Todas las de esa clínica | Solo la propia | Solo las asignadas |
| Crear, reprogramar o cancelar citas | En las agendas de su clínica, según las reglas de la operación | Solo en su agenda | Solo en agendas asignadas |
| Marcar `completed` y `no_show` | En las agendas de su clínica | No por este rol | Solo en agendas asignadas |
| Gestionar series recurrentes y sus ocurrencias | En las agendas de su clínica | Solo en su agenda | Solo en agendas asignadas |
| Cancelar por urgencia sin aprobación del paciente | En su clínica, con motivo y auditoría | No concedido a este rol en el diseño actual | Solo agendas asignadas, sin aprobación previa del administrador, con motivo y auditoría |
| Cambiar disponibilidad o bloqueos locales | De profesionales habilitados allí | Solo los propios | No inicialmente |
| Registrar y gestionar ausencia general | No por ser administrador local | Solo la propia | No concedido para otra persona |
| Buscar y consultar datos administrativos de pacientes | Los de esa clínica | Solo pacientes asignados explícitamente | Los de esa clínica, con respuesta administrativa limitada |
| Registrar pacientes y editar sus datos administrativos | Sí, dentro de esa clínica | No concedido inicialmente por este rol | Sí, dentro de esa clínica |
| Gestionar membresías, roles y asignaciones de agendas | Sí, dentro de esa clínica | No | No |
| Cambiar configuración de la clínica o tipos de cita | Sí, dentro de esa clínica | No | No |
| Habilitar tipos de cita para un profesional | Sí, dentro de esa clínica | No | No |
| Desactivar a un profesional globalmente | No por tener este rol local | No por tener este rol | No |
| Consultar notas o expediente clínico | No implícitamente | No implícitamente | No |

La asignación de pacientes al profesional todavía no existe en el código. Se
debe definir antes de habilitar ese permiso; mientras tanto se deniega. La mera
pertenencia a la misma clínica no permite al profesional ver todos los pacientes.
El expediente tendrá su política independiente cuando se implemente.

Para recepción, gestionar agenda incluye consultar citas, crear, reprogramar o
cancelar reservas individuales o recurrentes y marcar `completed` o `no_show`,
siempre dentro de sus agendas asignadas y de las reglas del dominio. No incluye
borrar citas físicamente, alterar duración configurada ni escribir notas clínicas.
Registrar asistencia no convierte a recepción ni al administrador en profesionales.

Una cancelación ordinaria requiere solicitud o aprobación del paciente. Recepción
puede registrar que la recibió por teléfono u otro canal operativo, guardando
canal, fecha y actor, sin exigir confirmación digital ni una cuenta del paciente.
No se requieren grabaciones ni conversaciones completas. El registro pertenece
a la misma cita y clínica y debe conservarse con su auditoría.

Solo ante urgencia del profesional se permite omitir esa aprobación. Administrador
y recepción pueden ejecutar la excepción dentro de su alcance; recepción no
necesita autorización previa del administrador. El rol profesional por sí solo
no recibe esta facultad en el diseño actual. La urgencia no amplía permisos ni
cambia a `no_show` una cita que el profesional no pudo atender.

La cancelación urgente deja aviso pendiente hasta registrar que se informó al
paciente, con canal, actor y fecha. Un intento sin contacto no lo marca como avisado.
No se espera aprobación o respuesta para cancelar, y registrar la cancelación no
significa que se haya enviado automáticamente un mensaje.

Crear un bloqueo o una ausencia es un permiso separado: recepción no obtiene
permiso general para bloquear horarios por poder cancelar. Una operación conjunta
de cancelación y bloqueo exige un actor autorizado para ambas y persistencia
atómica; no convierte en obligatoria la aprobación del administrador para cada
cancelación individual que recepción puede realizar por sí misma.

Las ausencias generales se representan en `professional_absences`, descrita en
el documento de agenda. Impiden nuevas reservas en todas las clínicas, pero no
cancelan citas existentes. Cada clínica resuelve las suyas con sus propios
permisos. La autoría acordada para la primera versión es el propio profesional;
no se concede facultad global a un administrador local ni se incluye delegación
para gestionar ausencias en nombre de otro profesional. Si un administrador es
también profesional, puede gestionar las propias por esa identidad. Para calcular
disponibilidad solo se comparte el intervalo no disponible, no motivos privados
ni datos de otras clínicas.

En operaciones recurrentes se debe seleccionar explícitamente una ocurrencia o
el conjunto de futuras citas afectadas. Cancelar una fecha no cancela toda la
serie. El sistema valida permisos sobre todas las agendas afectadas y registra
actor, motivo y alcance; no existe una autorización masiva por el simple hecho
de conocer el identificador de la serie.
Cambiar el profesional de una sola ocurrencia no cambia las otras; cambiarlo en
serie actualiza las futuras seleccionadas, incluso las todavía no generadas, sin
alterar citas finalizadas. En ambos casos se exige acceso a las agendas de origen
y destino, y el nuevo profesional debe estar habilitado para ese tipo de cita.

Recepción tiene permisos de registro y edición de pacientes independientes de
las asignaciones de agendas. Como alcance inicial propuesto, puede buscar,
consultar, registrar y editar datos administrativos de pacientes de su clínica,
incluso antes de que tengan una cita. Esto permite dar de alta a una persona sin
necesitar previamente una asignación a un profesional.

Los campos editables serán una lista explícita: nombre, apellidos, fecha de
nacimiento, teléfono y correo. No incluye notas, diagnósticos ni expedientes;
tampoco permite cambiar `id`, trasladar el paciente a otra clínica mediante
`clinic_id`, borrarlo, desactivarlo ni modificar campos internos de auditoría.
Agregar un campo al modelo de persistencia no lo vuelve editable automáticamente.

El backend determina la clínica del alta desde la ruta autorizada. Las lecturas
y actualizaciones filtran por paciente y clínica; conocer un ID no permite
operar sobre pacientes de otra clínica. Las respuestas para recepción contienen
solamente datos administrativos permitidos, no el objeto completo de persistencia.

Buscar un paciente de la clínica no concede acceso a sus citas en agendas no
asignadas ni a información clínica. Una membresía activa de recepción puede
gestionar datos administrativos aunque no tenga agendas asignadas; para gestionar
citas siempre necesita la asignación correspondiente. Desactivar la membresía
o retirar el rol elimina también estos permisos sobre pacientes.

La implementación deberá registrar quién creó o actualizó el registro y cuándo,
con datos de auditoría establecidos por el servidor, sin volcar datos personales
ni tokens en logs generales.

Los roles se combinan de forma aditiva dentro de una misma clínica: una persona
con `admin` y `receptionist` conserva el alcance administrativo. Una persona con
`professional` y `receptionist` puede gestionar su agenda y las otras agendas
asignadas. Ningún rol concede acceso implícito a otra clínica o al expediente.

## Reglas de seguridad y ciclo de vida

1. Obtener el usuario de un token autenticado, nunca de un `user_id` enviado como
   prueba de identidad en el body o de un encabezado inventado por el cliente.
2. Comprobar perfil de usuario activo, clínica activa, membresía activa y rol
   requerido. Un fallo al consultar permisos debe denegar, no permitir.
3. Resolver siempre la membresía para la clínica del endpoint. El rol en A no
   autoriza a consultar B aunque el usuario conozca sus identificadores.
4. Para crear o reprogramar citas, comprobar además que el titular está activo,
   habilitado en esa clínica y autorizado para el tipo de cita. El alcance se
   permite por rol administrativo, por titularidad con rol profesional, o por rol
   de recepción con asignación activa. Cerrar o cancelar citas existentes no exige
   que el titular o el tipo sigan activos; sí exige identidad, membresía y permisos
   vigentes del actor. Los permisos clínicos no se heredan de la gestión administrativa.
5. No confiar en un rol enviado por el usuario ni en metadatos que pueda editar.
   Consultar los roles actuales permite que una revocación tenga efecto sin
   esperar a que venza un token que conserve permisos antiguos.
6. Un administrador puede conceder o retirar roles en su clínica. La primera
   membresía administrativa se crea mediante un flujo controlado, nunca mediante
   autoasignación pública. Inicialmente puede aprovisionarse con una operación
   administrativa restringida. Un futuro alta de clínica deberá crear clínica y
   primera membresía juntas después de autenticar al creador.
7. No retirar ni desactivar al último administrador activo de una clínica activa.
   Esa comprobación requiere una transacción con control de concurrencia; una
   consulta seguida de otra petición independiente no es suficiente.
8. Desactivar una membresía bloquea acceso y nuevas reservas, pero no borra
   citas históricas. Las citas futuras existentes requieren revisión explícita;
   no se cancelan automáticamente como efecto secundario oculto. Ante una urgencia
   del profesional pueden cancelarse sin aprobación del paciente mediante el flujo
   explícito auditado del dominio de agenda, no como consecuencia implícita de desactivar.
9. Restringir el borrado de perfiles y relaciones referenciados por historial.
   La gestión de eliminación de cuentas Auth necesitará un proceso explícito;
   no aplicar cascadas que borren datos de atención accidentalmente.
10. Los datos de profesionales visibles para un administrador se limitan a los
    necesarios para su clínica; no exponer otras membresías de la persona.
11. Revisar asignaciones de recepción tanto en listados como al consultar o
    modificar una cita por ID. Al reprogramar hacia otro profesional, comprobar
    acceso a la agenda de origen y destino; no basta con autorizar la cita antigua.
12. Revocar una asignación quita acceso a esa agenda sin afectar a las otras.
    Retirar el rol `receptionist` o desactivar una de las membresías desactiva sus
    asignaciones relacionadas en la misma operación. Reactivar la membresía o
    el rol no recupera automáticamente las asignaciones: requiere autorización
    administrativa explícita. No se borran las citas afectadas.
    Si se desactivó la membresía del profesional, recepción deja de tener una
    asignación activa y el administrador debe resolver las citas pendientes; la
    urgencia no reactiva por sí sola permisos revocados.
13. `completed` y `no_show` pueden ser registrados por administrador o recepción
    solo en agendas autorizadas y mediante transiciones válidas. Las correcciones
    de estados finales requieren un flujo explícito, no un PATCH arbitrario.
    El rol profesional por sí solo no permite estos cierres. Tener además otro
    rol autorizado permite realizarlos dentro del alcance de ese otro rol.
    El cierre es manual y debe respetar los umbrales de tiempo del documento de
    agenda: para sesiones de 60 minutos, `completed` desde los 45 minutos del
    inicio y `no_show` desde los 30; los límites para sesiones menores siguen
    pendientes. Ningún rol administrativo permite ignorar esos umbrales.
    Una cita `scheduled` cuyo final ya pasó aparece como pendiente de cierre en
    las agendas autorizadas de administrador y recepción; no cambia automáticamente
    de estado ni se elimina. El actor registra el resultado real y la fecha de cierre.
14. No exigir perfil profesional al actor administrativo. Diferenciar siempre
    el usuario que gestiona la cita del profesional que atiende al paciente.
15. Una urgencia debe guardar su categoría, motivo administrativo, actor y fecha.
    Informar al paciente no es lo mismo que pedirle aprobación; se registra
    el aviso pendiente/realizado por el canal operativo, sin afirmar que se enviaron
    notificaciones automáticas. Un bloqueo de ausencia puede seguir impidiendo
    reservar un intervalo aunque su cita anterior esté cancelada.
16. Un bloqueo local afecta a una clínica y una ausencia general al profesional
    en todas sus clínicas. El alcance global de la indisponibilidad no concede
    privilegios globales para consultar pacientes, cancelar citas o ver motivos
    privados. Las citas existentes se resuelven explícitamente por cada clínica.

Una respuesta `401` indica falta de autenticación válida. Un `403` indica que
la identidad autenticada no tiene acceso a la operación o clínica solicitada.
Buscar registros siempre dentro del alcance autorizado permite responder `404`
cuando no están allí sin revelar su existencia en otra clínica.

## Cómo encaja en Go

Mantener la separación de responsabilidades del proyecto y agregar una capa de
servicio donde haya reglas que involucren varias entidades:

- **Middleware de autenticación:** valida la sesión y establece el usuario.
- **Autorización:** resuelve membresía, roles y alcance sobre la agenda solicitada.
- **Handler:** interpreta y valida el formato del request y devuelve HTTP.
- **Servicio:** aplica reglas, por ejemplo asignar rol profesional o conservar
  al último administrador activo.
- **Store:** ejecuta consultas y operaciones atómicas de persistencia.

`DBClient` seguirá siendo infraestructura compartida; no guardará al usuario
actual, clínica actual ni roles en variables globales. Esos datos pertenecen a
cada petición. Tampoco se debe cambiar el token de un cliente Supabase compartido
entre peticiones de distintos usuarios.

No exponer escrituras genéricas sobre las tablas de roles: saltarían las reglas
del servicio. Si se utiliza la Data API de Supabase, una operación que modifica
varias tablas atómicamente puede requerir una función de base de datos restringida,
en lugar de varias peticiones HTTP independientes.

## Seguridad pendiente en el proyecto actual

En la revisión local, `backend/cmd/main.go` registra los módulos de pacientes y
clínicas sin middleware de autenticación ni autorización. Filtrar por
`clinic_id` no demuestra que quien llama tenga permiso para esa clínica.

Antes de exponer datos reales hay que proteger también esos endpoints existentes.
No basta con proteger los nuevos módulos de membresías.

RLS y privilegios de tablas deberán definirse junto con la migración. No crear
políticas públicas permisivas para hacer funcionar las pruebas. Las claves
administrativas de Supabase pueden omitir RLS: su uso en el backend no sustituye
la autorización por usuario. Ninguna clave privilegiada debe llegar al frontend.
Todavía no se ha inspeccionado ni modificado la configuración remota de seguridad.

## Secuencia de implementación

La secuencia, las dependencias, los estados y la planificación se mantienen en el
[plan único de Notion](https://app.notion.com/p/46deaaf04e474ebab434087ae36be978).
Consultar las vistas **Secuencia** y **Para planificar**; este documento conserva
las reglas y los criterios técnicos, sin una segunda lista de trabajo.

Se puede avanzar en reglas puras sin construir antes toda la administración de
usuarios. La autenticación y autorización siguen siendo obligatorias antes de
exponer datos reales, incluidos los endpoints existentes de pacientes y clínicas.
Las migraciones reportadas como aplicadas deben verificarse y versionarse (DAT-01).

Pruebas de aceptación de usuarios, roles y su integración con agenda:

- Sin autenticación no se puede leer ni modificar información de clínicas.
- Un administrador de A no puede administrar B sin su respectivo rol allí.
- Un usuario puede combinar roles y ejercerlos solo en la membresía adecuada.
- Desactivar la membresía A no cambia el acceso a B.
- No se pueden crear membresías ni roles duplicados.
- No se puede asignar `professional` sin perfil profesional válido.
- Un dueño sin perfil profesional puede tener rol `admin` y gestionar su clínica.
- Un administrador sin perfil profesional puede marcar `completed` y `no_show`.
- Un profesional no puede autoasignarse `admin`.
- Un rol enviado en el request o en metadatos editables no concede acceso.
- Dos retiros simultáneos no pueden dejar una clínica sin administrador activo.
- Un error de consulta de permisos nunca concede acceso.
- Un administrador no obtiene permiso clínico ni permisos globales por accidente.
- Recepción sin asignaciones no puede acceder a ninguna agenda.
- Recepción puede registrar un paciente sin citas ni profesional asignado y
  editar únicamente sus campos administrativos permitidos dentro de su clínica.
- Una membresía de recepción sin agendas asignadas conserva esos permisos sobre
  pacientes, pero no obtiene acceso a citas por consultar los datos del paciente.
- Recepción no puede leer ni editar pacientes de otra clínica aunque conozca su ID.
- Intentar modificar clínica, identificador, auditoría o campos clínicos se rechaza.
- Recepción no puede borrar ni desactivar pacientes con estos permisos.
- Desactivar la membresía o retirar el rol de recepción revoca sus permisos de
  pacientes, salvo los que conserve por otro rol activo en esa misma clínica.
- Recepción puede gestionar varias agendas asignadas, pero no otras de la clínica.
- Recepción puede marcar `completed` y `no_show` solo en agendas asignadas, sin
  adquirir permisos sobre notas o expedientes clínicos.
- Un profesional sin rol administrativo no puede marcar `completed` ni `no_show`;
  un usuario con rol adicional autorizado sí puede hacerlo dentro de su alcance.
- Ningún rol permite crear o reprogramar citas hacia el pasado ni insertar altas
  retroactivas enviando un estado final. El mero paso del tiempo no decide asistencia.
- Cambiar profesional, día y hora en una ocurrencia o en sus futuras seleccionadas
  exige permisos en origen y destino y no altera otras ocurrencias ni el historial.
- El permiso para gestionar una serie se limita a las agendas autorizadas; no
  permite modificar ocurrencias o series de otras clínicas.
- Solo el administrador habilita tipos por profesional y no puede mezclar clínicas
  al crear esa relación. Pertenecer a una clínica no habilita todos sus servicios.
- La cancelación urgente sin aprobación requiere motivo y auditoría; la ordinaria
  no puede usar esa excepción. La urgencia no restaura permisos revocados.
- Recepción puede cancelar por urgencia en una agenda asignada sin aprobación
  previa del administrador y no puede hacerlo en agendas ajenas.
- Recepción registra una solicitud ordinaria por teléfono con canal, fecha y actor,
  sin confirmación digital ni acceso a notas clínicas.
- Registrar una cancelación no marca al paciente como avisado; un aviso realizado
  requiere registro explícito. Un intento fallido conserva el aviso pendiente.
- Administrador y recepción pueden resolver citas vencidas pendientes de cierre
  dentro de su alcance; el sistema no decide automáticamente que hubo inasistencia.
- Un bloqueo de A no impide reservar en B; una ausencia general sí impide nuevas
  reservas en ambas sin permitir a A leer o cancelar citas de B.
- Cancelar una ocurrencia no cancela las siguientes ni las regenera en un reintento.
- Conocer el ID de una cita no permite saltarse la autorización de su agenda.
- No es posible asignar una agenda de una clínica a una membresía de otra.
- Recepción no puede conceder roles, autoasignarse agendas ni modificar configuración.
- Revocar una asignación bloquea acceso sin afectar otras agendas ni borrar citas.
- Reprogramar hacia una agenda no asignada se rechaza aunque la de origen lo esté.
- Desactivar y reactivar una membresía no restaura asignaciones revocadas.
- Un usuario con varios roles obtiene solo la unión de sus alcances autorizados
  en la clínica consultada, sin privilegios globales ni acceso clínico implícito.

## Referencias de la propuesta Supabase

- [Datos propios de usuarios y referencia a Auth](https://supabase.com/docs/guides/auth/managing-user-data).
- [RLS y claves que omiten estas políticas](https://supabase.com/docs/guides/database/postgres/row-level-security).

Estas fuentes sustentan la integración técnica con Supabase; el modelo de roles,
las reglas de clínica y las decisiones de alcance de este documento son propuestas
de diseño para este proyecto.
