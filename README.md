# Тестовое задание для стажера-разработчика

Реализовать систему для добавления и чтения постов и комментариев с использованием GraphQL, аналогичную комментариям к постам на популярных платформах, таких как Хабр или Reddit.

Характеристики системы постов:
* Можно просмотреть список постов;
* Можно просмотреть пост и комментарии под ним.;
* Пользователь, написавший пост, может запретить оставление комментариев к своему посту.

Характеристики системы комментариев к постам:
* Комментарии организованы иерархически, позволяя вложенность без ограничений;
* Длина текста комментария ограничена до, например, 2000 символов;
* Система пагинации для получения списка комментариев;

(*) Дополнительные требования для реализации через GraphQL Subscriptions:

* Комментарии к постам должны доставляться асинхронно, т.е. клиенты, подписанные на определенный пост, должны получать уведомления о новых комментариях без необходимости повторного запроса.

Требования к реализации:
* Система должна быть написана на языке Go;
* Использование Docker для распространения сервиса в виде Docker-образа;
* Хранение данных может быть как в памяти (in-memory), так и в PostgreSQL. Выбор хранилища должен быть определяемым параметром при запуске сервиса;
* Покрытие реализованного функционала unit-тестами;

Критерии оценки:
* Как хранятся комментарии и как организована таблица в базе данных/in-memory, включая механизм пагинации.
* Качество и чистота кода, структура проекта и распределение файлов по пакетам.
* Обработка ошибок в различных сценариях использования.
* Удобство и логичность использования системы комментариев.
* Эффективность работы системы при множественном одновременном использовании, сравнимая с популярными сервисами, такими как Хабр.
* В реализации учитываются возможные проблемы с производительностью, такие как проблемы с n+1 запросами и большая вложенность комментариев.


## Комментарий к реализации

Сервис даёт возможность сохранять посты и комментарии, получать посты и комментарии к ним с пагинацией.
Limit для пагинации определён в config файле.

Сервис даёт возможность подписаться на новые комментарии по id поста через GraphQL Subscriptions.

Для пагинации комментариев было решено использоватт метод курсора.
Так как не было требований к сортировке комменатриев по рейтингу, все комментарии сохраняются и отдаются в отсортированном по дате создания виде.
Иерархичность не ограничена.
Длина комментариев ограничена переменной, которая определяется в config файле.

Решено реализовать in-memory решение без сохранения данных между запусками сервера.
В качестве хранилища используется `slice`. Индекс элемента выступает в роли его id.


В случае, если необходимо сохранять состояние между запусками сервера или запустить несколько экземпляров сервиса, имеет
смысл использовать для хранения `Redis`.

Проблема N + 1 решена двумя dataloaders, благодаря чему при запросе поста и комментариев к нему (в т.ч. вложенных) происходит 3 запроса в БД.

Вложенность комментариев в запросе ограничена переменной, которая определяется в app.go файле.

Требует доработки:
* Сервис необходимо покрыть unit-тестами.
* Написать docker-файл для запуска в контейнере.

## Общее описание решения

- Сервис реализован на языке `Golang` версии `1.22`.
- Для работы с raphQL используется библиотека [gqlgen](https://github.com/99designs/gqlgen).
- В качестве СУБД используется `PostgreSQL`. В качестве библиотеки для работы с запросами к `PostgreSQL` используется
  [pgxpool](https://github.com/jackc/pgx).

Существует три режима работы:
* `envLocal`
* `envDev`
* `envProd`
