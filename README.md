<h1 align="center"> Привет! Я <a target="_blank"> Кармеев Артур из группы ЭФМО-01-25 </a> 
<img src="https://github.com/blackcater/blackcater/raw/main/images/Hi.gif" height="32"/></h1>
<h3 align="center"> Данная практика была интересной 🤔 </h3>

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













