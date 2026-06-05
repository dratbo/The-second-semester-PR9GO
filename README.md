<h1 align="center"> Привет! Я <a target="_blank"> Кармеев Артур из группы ЭФМО-01-25 </a> 
<img src="https://github.com/blackcater/blackcater/raw/main/images/Hi.gif" height="32"/></h1>
<h3 align="center"> Данная практика была не из лёгких :face_with_head_bandage: </h3>

<h3 align="center"> Практическая работа №9: Реализация распределённого кэша (Redis cluster) </h3>


Структура работы:

    └── pz9-redis-cache/
        ├── go.mod
        ├── go.sum
        ├── README.md
        ├── .idea/
        │   ├── .gitignore
        │   ├── modules.xml
        │   ├── pz9-redis-cache.iml
        │   ├── vcs.xml
        │   └── workspace.xml
        ├── .git/
        ├── internal/
        │   ├── task/
        │   │   ├── model.go
        │   │   └── repo.go
        │   ├── service/
        │   │   └── task_service.go
        │   ├── httpapi/
        │   │   └── handler.go
        │   ├── config/
        │   │   └── config.go
        │   └── cache/
        │       ├── keys.go
        │       ├── redis.go
        │       └── ttl.go
        ├── deploy/
        │   └── redis/
        │       └── docker-compose.yml
        └── cmd/
            └── server/
                └── main.go


## 1. Немного теории

Почему Redis подходит для кэширования
Redis — это очень быстрый in-memory key-value store. Его удобно использовать как внешний кэш, потому что:
•	доступ к данным быстрый;
•	можно задавать TTL;
•	можно удалять и обновлять ключи выборочно;
•	Redis легко поднять локально в Docker;
•	он хорошо подходит для cache-aside сценариев.
Для учебной практики Redis удобен ещё и тем, что его поведение легко наблюдать: можно увидеть hit, miss, set, del и момент истечения TTL.

Что такое распределённый кэш
Распределённым называют кэш, который существует не внутри одного процесса приложения, а как отдельная инфраструктурная система, доступная по сети. В отличие от обычной map в памяти приложения, такой кэш:
•	не привязан к одному экземпляру сервиса;
•	может использоваться несколькими экземплярами приложения;
•	может быть развернут как отдельный узел или кластер.
В данной работе используется учебный стенд Redis cluster или приближённая конфигурация. Это позволяет воспринимать Redis не как вспомогательную локальную переменную, а как внешнюю зависимость backend-системы.

Что такое cache-aside
Стратегия cache-aside — это один из самых распространённых подходов к кэшированию чтения.
Смысл алгоритма:
1.	Сначала приложение пытается получить данные из кэша.
2.	Если данные найдены — они возвращаются клиенту.
3.	Если данных нет — приложение идёт в БД.
4.	Если БД вернула данные — приложение кладёт их в кэш.
5.	Затем ответ возвращается клиенту.
Это означает, что приложение само управляет наполнением кэша. Redis не подменяет БД и не знает бизнес-логику приложения. Именно сервис решает, когда читать из кэша, когда обращаться в БД и когда обновлять ключ.

Почему Redis не является источником истины
Очень важно понимать архитектурную роль Redis.
Redis в данном ПЗ:
•	не хранит канонические данные;
•	не определяет, существует ли задача на самом деле;
•	не гарантирует долговременное хранение;
•	не должен ломать API при недоступности.
Источником истины остаётся БД или основной репозиторий. Redis нужен только для ускорения повторного чтения.
Это принципиальный момент. Если Redis внезапно недоступен, сервис всё равно должен обслуживать запросы через основное хранилище.

Что такое TTL
TTL — это time to live, то есть время жизни ключа в кэше.
Если TTL не использовать, данные могут храниться бесконечно долго, а значит:
•	кэш переполняется;
•	данные устаревают;
•	приложение всё сильнее зависит от ручной инвалидации.
Для учебной работы разумны такие значения:
•	для одной сущности: 60–300 секунд;
•	для списка: обычно меньше, если список часто меняется.
TTL позволяет кэшу самоочищаться и со временем обновлять содержимое.

Что такое jitter и зачем он нужен
Если всем ключам назначить одинаковый TTL, то множество записей может истечь почти одновременно. Тогда большое число запросов резко пойдёт не в Redis, а в БД. Это создаёт всплеск нагрузки.
Чтобы уменьшить этот эффект, к базовому TTL добавляют небольшой случайный разброс — jitter.
Пример:
•	базовый TTL = 120 секунд;
•	jitter = от 0 до 30 секунд;
•	итоговый TTL = 120 + случайное число от 0 до 30.
Так ключи истекают не одновременно, а более равномерно.


## 2. Запуск упрощенного варианта: один Redis в docker compose

<table cellpadding="10">
  <tr>
    <td><img width="974" height="338" alt="image" src="https://github.com/user-attachments/assets/5deef5a5-5f25-4c12-9c9f-a3a35a38b1e9" /></td>
  </tr>
</table>

## 3. Проверка чтения и заполнения кэша

Запускаем шарманку

<table cellpadding="10">
  <tr>
    <td><img width="974" height="364" alt="image" src="https://github.com/user-attachments/assets/860017c4-7f88-4bd1-8fb1-33cfa0eb3bf9" /></td>
  </tr>
</table>

Делаем первый запрос

<table cellpadding="10">
  <tr>
    <td><img width="974" height="512" alt="image" src="https://github.com/user-attachments/assets/aae0997e-1eb3-400f-869e-836d1257c358" /></td>
  </tr>
</table>

Сразу второй

<table cellpadding="10">
  <tr>
    <td><img width="974" height="513" alt="image" src="https://github.com/user-attachments/assets/a05aaaeb-c0e5-4c60-8da3-1e5e0a7bc951" /></td>
  </tr>
</table>

По логам видим, что сначала у нас `cache miss`, тоесть мы взяли данные на прямую из БД, а при втором запросе у нас `cache hit`, следовательно при втором запросе данные были взяты из кэша! 🤠

<table cellpadding="10">
  <tr>
    <td><img width="974" height="254" alt="image" src="https://github.com/user-attachments/assets/7aa8e77d-6e82-4735-b4b1-787d75c4e585" /></td>
  </tr>
</table>


## 4. Проверка инвалидации при изменении

Меняем задачу!

<table cellpadding="10">
  <tr>
    <td><img width="974" height="511" alt="image" src="https://github.com/user-attachments/assets/e864b515-7c0f-4841-8bd0-a8e29d1a1dc8" /></td>
  </tr>
</table>

Сразу делаем запрос!

<table cellpadding="10">
  <tr>
    <td><img width="974" height="513" alt="image" src="https://github.com/user-attachments/assets/8d8842a7-29f9-4ae5-a6b3-c4cd3e05e65d" /></td>
  </tr>
</table>

По логам видим, после изменения задачи старые данные удалились из кэша, значит все опять работает! :smirk:

<table cellpadding="10">
  <tr>
    <td><img width="974" height="283" alt="image" src="https://github.com/user-attachments/assets/97054205-4e9f-43ea-9af1-b8aaae05eae5" /></td>
  </tr>
</table>


## 5. Проверка удаления

Удаляем задачу!

<table cellpadding="10">
  <tr>
    <td><img width="974" height="511" alt="image" src="https://github.com/user-attachments/assets/34e267ed-1759-49ca-b4e1-caecb9bf4ce7" /></td>
  </tr>
</table>

Всё удалено успешно!

<table cellpadding="10">
  <tr>
    <td><img width="974" height="512" alt="image" src="https://github.com/user-attachments/assets/b4482e72-81f2-4795-a92a-8102cb83933d" /></td>
  </tr>
</table>


## 6. Проверка деградации при остановке Redis

Останавливаем наш не истинный источник Redis :anguished:

<table cellpadding="10">
  <tr>
    <td><img width="974" height="95" alt="image" src="https://github.com/user-attachments/assets/f86d9d20-284a-4112-a229-939296628dfc" /></td>
  </tr>
</table>

Делаем запрос

<table cellpadding="10">
  <tr>
    <td><img width="974" height="515" alt="image" src="https://github.com/user-attachments/assets/7c92c677-2686-4e7c-b379-4cf262739caa" /></td>
  </tr>
</table>

Запрос прошел успешно, как и ожидалось. В логах нам дают знать, что с нашим Redis что то случилось :cry:

<table cellpadding="10">
  <tr>
    <td><img width="974" height="260" alt="image" src="https://github.com/user-attachments/assets/b5723604-e785-490b-9675-2da71bf29756" /></td>
  </tr>
</table>


## 7. Доп задание

---

Вариант 1. Кэширование списка задач
Реализуйте кэширование маршрута:
GET /v1/tasks
с ключом:
tasks:list
или с учётом параметров пагинации.

Не буду врать, здесь я запутался сильно, взял небольшую помощь у гпт :construction_worker:

---

Пройдемся чутка по коду, что поменял, что добавил

Фиксированный ключ для полного списка без query-параметров

<table cellpadding="10">
  <tr>
    <td><img width="974" height="519" alt="image" src="https://github.com/user-attachments/assets/06d359eb-27e1-4660-9188-df7db884ee92" /></td>
  </tr>
</table>

Нужен для стабильного порядка задач в списке (по id), иначе порядок из map каждый раз разный 

<table cellpadding="10">
  <tr>
    <td><img width="974" height="516" alt="image" src="https://github.com/user-attachments/assets/fb3ad0d9-ab5b-45fc-ba77-da2d28e595d1" /></td>
  </tr>
</table>

Метод `GetTasksList` работает как cache-aside: сначала читает JSON из Redis по ключу `tasks:list` или `tasks:list:page=…:limit=…`, при промахе берёт данные из репозитория и записывает их в кэш с более коротким TTL, чем у одной задачи.
При PATCH и DELETE вызывается `invalidateTasksListCache`: удаляются все ключи `tasks:list*`, чтобы список в кэше не оставался устаревшим после изменения задачи. 
Тут больше всего мучался. 

<table cellpadding="10">
  <tr>
    <td><img width="974" height="508" alt="image" src="https://github.com/user-attachments/assets/bf5df245-75b0-4943-888d-9f31957a1cc1" /></td>
  </tr>
</table>


Добавлена целая функция в handlers.go! HTTP-слой не знает про Redis; только парсит запрос и вызывает сервис — разделение ответственности (handler / service / cache / repo)


<table cellpadding="10">
  <tr>
    <td><img width="974" height="519" alt="image" src="https://github.com/user-attachments/assets/8051ea8b-dc08-4355-82ca-9cc8fe44058a" /></td>
  </tr>
</table>

Ну и в main.go прописал новые маршруты

---

### Результаты:

Всё запустили и поехали! По стандарту делаем два запроса


<table cellpadding="10">
  <tr>
    <td><img width="974" height="509" alt="image" src="https://github.com/user-attachments/assets/38aaf1a0-91dc-4d2a-95be-405ce2dc3841" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="974" height="515" alt="image" src="https://github.com/user-attachments/assets/b9003423-4961-41fe-8caa-f8454d0a0e7b" /></td>
  </tr>
</table>

В логах: первый раз `cache miss: tasks:list`, второй — `cache hit: tasks:list`. Всё сработало!

<table cellpadding="10">
  <tr>
    <td><img width="974" height="393" alt="image" src="https://github.com/user-attachments/assets/69eb16db-3537-457c-9462-d97f87c1c079" /></td>
  </tr>
</table>

Теперь с пагинацией :disguised_face:



<table cellpadding="10">
  <tr>
    <td><img width="974" height="512" alt="image" src="https://github.com/user-attachments/assets/8fe63bca-c4a9-4cfe-a956-d986c442a9c6" /></td>
  </tr>
</table>

В логах: `cache miss: tasks:list:page=1:limit=10` → затем `cache hit`.


<table cellpadding="10">
  <tr>
    <td><img width="949" height="170" alt="image" src="https://github.com/user-attachments/assets/e92c956f-c8e5-47ef-97ab-49d22a7377e3" /></td>
  </tr>
</table>

Инвалидация (после PATCH по id)

<table cellpadding="10">
  <tr>
    <td><img width="974" height="513" alt="image" src="https://github.com/user-attachments/assets/5e4a10a0-2ba4-4a09-8cc2-286404515a72" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="822" height="52" alt="image" src="https://github.com/user-attachments/assets/a3cf82e9-d812-4f8d-9839-9c9e4cd308bd" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="974" height="515" alt="image" src="https://github.com/user-attachments/assets/eef39923-5799-4978-b1a4-b889f334fae1" /></td>
  </tr>
</table>

В логах: `cache invalidated: tasks:list*`, затем снова `cache miss` для списка

<table cellpadding="10">
  <tr>
    <td><img width="974" height="566" alt="image" src="https://github.com/user-attachments/assets/753a3e39-63a4-41d9-82c4-bd3a60cc62db" /></td>
  </tr>
</table>


## 8. Контрольные вопросы :exploding_head:


1. Что такое cache-aside?

Cache-aside — стратегия, при которой приложение само управляет кэшем: сначала читает Redis; при попадании (hit) отдаёт данные из кэша; при промахе (miss) читает основное хранилище (БД/репозиторий), кладёт результат в Redis и отвечает клиенту. Redis не ходит в БД сам — логика в сервисе.


2. Почему Redis не должен быть источником истины?

Источник истины — основное хранилище (PostgreSQL, репозиторий и т.д.): там данные сохраняются надёжно и полно. Redis — кэш в памяти: может очиститься, упасть, истечь по TTL. Если считать Redis «главной БД», при сбое или потере кэша теряются или искажаются данные.


3. Зачем нужен TTL?

TTL (time to live) — время жизни ключа в кэше. Без TTL устаревшие данные могут храниться слишком долго после изменений в БД. TTL ограничивает «возраст» кэша и снижает риск долго отдавать неактуальную информацию, даже если инвалидацию забыли.


4. Что такое jitter?

Jitter — случайная добавка к TTL (например, база 120 с + от 0 до 30 с). Нужен, чтобы ключи не истекали одновременно: иначе много запросов разом пойдут в БД (cache stampede / thundering herd).


5. Почему одинаковый TTL для всех ключей может быть проблемой?

Если у многих ключей TTL заканчивается в один момент, кэш «обнуляется» пачкой и нагрузка на БД резко растёт. Jitter разносит моменты истечения во времени и сглаживает пики нагрузки.


6. Как должен вести себя сервис при недоступности Redis?

Сервис не должен падать из‑за Redis: ошибки чтения/записи логируются, запрос обрабатывается через репозиторий/БД (fallback). Клиент получает ответ, если данные есть в основном хранилище; кэш при недоступности Redis просто не используется. В вашей практике при старте — предупреждение в лог, при запросе — работа без кэша.


7. Почему кэш нужно инвалидировать после изменения данных?

После PATCH/DELETE в БД данные уже другие, а в Redis может остаться старая запись. Без инвалидации (удаления или обновления ключа) клиент получит устаревший ответ (cache hit по «грязным» данным). Поэтому после изменения сбрасывают `tasks:task:<id>` и при необходимости ключи списка `tasks:list*`.


8. Чем кэширование одной сущности проще, чем кэширование списка?

Одна задача — один ключ, одна запись, простая инвалидация при изменении этой задачи. Список зависит от состава всех задач: любое изменение/удаление может сделать список в кэше неверным; нужны отдельные ключи (в т.ч. с пагинацией) и сброс нескольких ключей. Списки в методичке кэшируют осторожнее или на меньший TTL.


9. В чём смысл ключа вида tasks:task:<id>?

Это предсказуемое имя кэша: префикс сервиса (`tasks`), тип объекта (`task`), идентификатор (`<id>`). Удобно искать, отлаживать, не смешивать с другими ключами (`tasks:list`), согласовано с API `GET /v1/tasks/{id}`.


10. Почему Redis рассматривается как внешняя инфраструктурная зависимость?

Redis — отдельный процесс/сервис (часто в Docker), не часть кода приложения. Он может быть недоступен, перезапущен, обновлён отдельно; от него зависят таймауты, сеть, конфигурация. Как и БД, его подключают снаружи — поэтому это инфраструктурная зависимость, а не встроенная библиотека в процессе API.
