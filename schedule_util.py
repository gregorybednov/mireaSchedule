#!/usr/bin/env python3
import openpyxl
#import shutil
import aiohttp
import asyncio
import re
import sys
from functools import reduce

csv_delimiter = ';'


async def urls():
    async with aiohttp.ClientSession() as s:
        async with s.get("https://mirea.ru/schedule") as r:
            if not r.ok:
                print("Ошибка сети. Сервер МИРЭА вернул код:", r.status, file=sys.stderr)
                return []
            t = await r.text()
            return list(re.findall("https://webservices.mirea.ru[^\"\']*II[TI][^\"\']*.xlsx", t))


async def filenames(urls):
    filenames = []
    for i, url in enumerate(urls):
        name = url[str.rfind(url, "/") + 1:]
        async with aiohttp.ClientSession() as s:
            async with s.get(url) as r:
                if r.ok:
                    with open(name, 'wb') as f:
                        t = await r.content.read()
                        f.write(t)
                else:
                        print("Ошибка сети. Сервер МИРЭА вернул код:", r.status, file=sys.stderr)
        filenames.append(name)
    return filenames


def complete_table(filename, status_string_update=None):
    wb = openpyxl.load_workbook(filename)
    ws = wb.active
    ws.delete_rows(88, ws.max_row)  # примечания не нужны

    cols = list(filter(
        lambda column: not(column[2].value is None or column[2].value.replace("\n", " ").replace("  ", " ") in ["№ пары", "Нач. занятий", "Оконч. занятий", "Неделя", "Ссылка"]),
        ws.columns))

    tables = dict(dict())
    current_group = ''
    result = []
    for col in cols:
        if col[2].value == "Дисциплина":
            current_group = col[1].value
            tables[current_group] = dict()
        tables[current_group][col[2].value] = list(
            map(lambda x: x.value.replace("\n", " ").replace("  ", " ") if isinstance(x.value, str) else x.value, col))[3:]

    for group in tables:
        for i in range(len(tables[group]['Дисциплина'])):
            if tables[group]['Дисциплина'][i] == '':
                continue
            newstr = csv_delimiter.join(
                [group, ("I" if (i % 2) != 1 else "II"), list(["пн", "вт", "ср", "чт", "пт", "сб"])[i // 14],
                 str((i % 14) // 2 + 1)])
            for record in tables[group]:
                if tables[group][record][i] is None:
                    newstr += csv_delimiter
                else:
                    newstr += csv_delimiter + tables[group][record][i]
            result.append(newstr)
    return result


async def make_allgroup():
    return reduce(
        lambda x, y: x + y,
        map(
            complete_table,
            await filenames(await urls())
        )
    )


if len(sys.argv) > 1:
    if sys.argv[1] == '--text-mode':
        print("\n".join(asyncio.run(make_allgroup())))
else:
    import PySimpleGUI as sg

    layout = [[sg.Text("Какой текст искать?")],
              [sg.Input(key='-INPUT-')],
              [sg.Text(size=(40, 1), key='-STATUS-')],
              [sg.Button('OK'), sg.Button('Выход')]]

    window = sg.Window('Поиск по расписанию', layout)

    while True:
        event, values = window.read()
        if event == sg.WINDOW_CLOSED or event == 'Выход':
            break

        if event == 'OK':
            if values["-INPUT-"] == '':
                window["-STATUS-"].update("Без фильтра будут выведены все занятия")
            allgroup_table = asyncio.run(make_allgroup())
            toout = []
            if values["-INPUT-"] == '':
                toout = allgroup_table
                filename = "EXPORT"
            else:
                toout = list(filter(lambda x: values["-INPUT-"] in x, allgroup_table))
                filename = values["-INPUT-"]
            if toout:
                with open(filename + ".csv", "w") as f:
                    f.write("\n".join(toout))
                window["-STATUS-"].update("Экспортировано в:" + filename + ".csv")
            else:
                window["-STATUS-"].update("Пары не найдены!")
    window.close()
