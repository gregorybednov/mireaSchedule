# mireaSchedule

Скрипт (точнее, теперь уже программа) для автоматического скачивания и парсинга расписания МИРЭА (по состоянию на 2023 год)
в текстовую и табличную (CSV) форму. (Допускается экспорт в формат HTML, корректно отображаемый в любом веб-браузере)

# Установка

Посмотрите релизы (Releases), чтобы скачать .exe файл для Windows под платформу x64.

Для GNU/Linux необходима ручная установка программы ```zenity``` для отображения диалоговых окон.

# Использование

При использовании графического интерфейса предлагается ввести строку текста.
Будут найдены занятия, имеющие данную строку в ячейках (это может быть фамилия
и инициалы преподавателя, аудитория в том же формате, что в и оригинальном файле
расписания, или официальное название предмета). 

Далее будет предложено сохранить файл с таким же названием как и запрос, но название в этот момент можно изменить.
Все найденные занятия списком экспортируются в текстово-табличном формате CSV c разделителем  **; (точка с запятой)**.
Просмотр можно настроить в Excel или Calc, в принципе файлы CSV можно открывать хоть в Блокноте.
На данный момент рекомендуемым способом вывода таблицы можно считать экспорт в HTML-таблицу.

Также запросы можно совмещать (по ИЛИ), и в таком случае будет выведено объединение множеств, соответствующих первому запросу и второму. Соединение происходит по символу **~ (тильда)**. Пример запроса: *Иванов В.А.~Петров А.Б.* выдаст все занятия, проводимые Ивановым В.А. и Петровым А.Б., а *Сергеев М.О.~ИКБО-99-24* выдаст пары, проводимые Сергеевым М.О. и пары, которые поставлены у ИКБО-99-24.

Если занятий не найдено, то будет выведено соответствующее сообщение и файл создаваться не будет.

Если в поле ввода ничего не вводить, то в таблицу будут выведены *все* занятия в расписаниях институтов ИТ и ИИ, но сохранить файл с пустым именем, вероятнее всего, вам не даст операционная система.

Также существует текстовый режим для работы в терминале, он вызывается аргументом командной строки ```--text-mode``` и не поддерживает фильтров (предполагается, что фильтры пользователь может реализовать через grep и подобные стандартные инструменты). В Windows данный режим не работает, пользуйтесь графическим интерфейсом или соберите программу самостоятельно через команду ```go build```

# Принцип работы

1. Утилита скачивает xlsx-файлы, помеченные как IIT, III и IRI (относящиеся к ИИТ, ИИИ и ИРИ) с официального сайта mirea.ru/schedule
2. Затем "конструирует" из человекочитаемой xlsx-таблицы (точнее, всех таблиц) длиннный "список" из всех существующих занятий.
3. После этого применяется поиск (точный, как в расписании) по записям "списка". При наличии нескольких запросов, совмещенных через тильду, удовлетворение записью в расписании хотя бы одного запроса считается достаточным, чтобы вывести их в итоговый файл.
4. Подходящие записи остаются в файл "текстовых таблиц" CSV, который затем, возможно, конвертируется в таблицу HTML
