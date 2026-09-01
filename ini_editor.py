# ini_editor.py
# INI Color Editor на Python

import sys
import os
import re
import argparse
from collections import OrderedDict

# ANSI-цвета (тёмная тема по умолчанию)
COLORS = {
    'reset': '\033[0m',
    'section': '\033[96m',     # cyan
    'key': '\033[92m',         # green
    'value': '\033[93m',       # yellow
    'comment': '\033[90m',     # gray
    'equals': '\033[37m',      # white
    'bracket': '\033[37m',     # white
}

LIGHT_COLORS = {
    'reset': '\033[0m',
    'section': '\033[94m',     # blue
    'key': '\033[32m',         # green
    'value': '\033[33m',       # yellow
    'comment': '\033[37m',     # white (на светлом фоне лучше серый, но используем)
    'equals': '\033[30m',      # black
    'bracket': '\033[30m',     # black
}

class INIFile:
    def __init__(self):
        self.sections = OrderedDict()  # section -> OrderedDict of key->value
        self.comments = {}  # хранение комментариев для сохранения (упрощённо)
        self.raw_lines = []  # для сохранения порядка и комментариев

    def parse(self, content):
        """Парсит содержимое INI-файла."""
        lines = content.splitlines()
        current_section = None
        self.raw_lines = lines[:]
        self.sections = OrderedDict()
        self.comments = {}
        for line in lines:
            line_stripped = line.strip()
            if not line_stripped or line_stripped.startswith(';') or line_stripped.startswith('#'):
                # комментарий или пустая строка – пропускаем (сохраним при записи)
                continue
            if line_stripped.startswith('[') and line_stripped.endswith(']'):
                section = line_stripped[1:-1].strip()
                if section not in self.sections:
                    self.sections[section] = OrderedDict()
                current_section = section
            elif '=' in line_stripped and current_section is not None:
                key, value = line_stripped.split('=', 1)
                key = key.strip()
                value = value.strip()
                self.sections[current_section][key] = value
        # Если нет секций, создаём секцию по умолчанию
        if not self.sections:
            self.sections[''] = OrderedDict()
        return self

    def get_value(self, section, key):
        return self.sections.get(section, {}).get(key)

    def set_value(self, section, key, value):
        if section not in self.sections:
            self.sections[section] = OrderedDict()
        self.sections[section][key] = value

    def delete_key(self, section, key):
        if section in self.sections and key in self.sections[section]:
            del self.sections[section][key]
            return True
        return False

    def delete_section(self, section):
        if section in self.sections:
            del self.sections[section]
            return True
        return False

    def add_section(self, section):
        if section not in self.sections:
            self.sections[section] = OrderedDict()
            return True
        return False

    def to_string(self, color=True, theme='dark'):
        """Возвращает строковое представление INI с подсветкой."""
        colors = LIGHT_COLORS if theme == 'light' else COLORS
        lines = []
        for section, keys in self.sections.items():
            if section:
                lines.append(f"{colors['bracket']}[{colors['reset']}{colors['section']}{section}{colors['reset']}{colors['bracket']}]{colors['reset']}")
            for key, value in keys.items():
                lines.append(f"{colors['key']}{key}{colors['reset']}{colors['equals']} = {colors['reset']}{colors['value']}{value}{colors['reset']}")
            if keys:
                lines.append('')
        return '\n'.join(lines)

    def save(self, filename):
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(self.to_string(color=False))

def main():
    parser = argparse.ArgumentParser(description='INI Color Editor')
    parser.add_argument('file', nargs='?', help='INI файл')
    parser.add_argument('-v', '--view', action='store_true', help='Показать содержимое')
    parser.add_argument('-s', '--set', help='Установить ключ (секция.ключ=значение)')
    parser.add_argument('-g', '--get', help='Получить значение ключа (секция.ключ)')
    parser.add_argument('-d', '--delete', help='Удалить ключ (секция.ключ)')
    parser.add_argument('-a', '--add', help='Добавить секцию')
    parser.add_argument('-r', '--remove', help='Удалить секцию')
    parser.add_argument('-o', '--output', help='Сохранить в другой файл')
    parser.add_argument('--color', action='store_true', default=True, help='Цветной вывод')
    parser.add_argument('--no-color', action='store_false', dest='color', help='Отключить цвета')
    parser.add_argument('--theme', choices=['light', 'dark'], default='dark', help='Цветовая тема')
    args = parser.parse_args()

    if not args.file:
        # Интерактивный режим
        filename = input("Введите путь к INI файлу: ").strip()
        if not filename:
            print("Файл не указан.")
            sys.exit(1)
        args.file = filename

    if not os.path.exists(args.file):
        print(f"Файл {args.file} не найден. Создаём новый.")
        ini = INIFile()
        ini.sections = OrderedDict()
    else:
        with open(args.file, 'r', encoding='utf-8') as f:
            content = f.read()
        ini = INIFile().parse(content)

    # Обработка команд
    if args.set:
        try:
            section_key, value = args.set.split('=', 1)
            if '.' in section_key:
                section, key = section_key.split('.', 1)
            else:
                section = ''
                key = section_key
            ini.set_value(section.strip(), key.strip(), value.strip())
            print(f"Установлено: {section}.{key} = {value}")
        except ValueError:
            print("Неверный формат --set. Используйте section.key=value")
            sys.exit(1)

    if args.get:
        try:
            if '.' in args.get:
                section, key = args.get.split('.', 1)
            else:
                section = ''
                key = args.get
            value = ini.get_value(section.strip(), key.strip())
            if value is not None:
                print(value)
            else:
                print(f"Ключ {args.get} не найден.")
        except:
            print("Неверный формат --get. Используйте section.key")
            sys.exit(1)

    if args.delete:
        try:
            if '.' in args.delete:
                section, key = args.delete.split('.', 1)
            else:
                section = ''
                key = args.delete
            if ini.delete_key(section.strip(), key.strip()):
                print(f"Ключ {args.delete} удалён.")
            else:
                print(f"Ключ {args.delete} не найден.")
        except:
            print("Неверный формат --delete. Используйте section.key")
            sys.exit(1)

    if args.add:
        if ini.add_section(args.add.strip()):
            print(f"Секция {args.add} добавлена.")
        else:
            print(f"Секция {args.add} уже существует.")

    if args.remove:
        if ini.delete_section(args.remove.strip()):
            print(f"Секция {args.remove} удалена.")
        else:
            print(f"Секция {args.remove} не найдена.")

    # Вывод
    if args.view or (not args.set and not args.get and not args.delete and not args.add and not args.remove):
        output = ini.to_string(color=args.color, theme=args.theme)
        if args.output:
            with open(args.output, 'w', encoding='utf-8') as f:
                f.write(ini.to_string(color=False))
            print(f"Сохранено в {args.output}")
        else:
            print(output)

    # Если были изменения и не указан output, сохраняем в исходный файл
    if (args.set or args.delete or args.add or args.remove) and not args.output:
        ini.save(args.file)
        print(f"Изменения сохранены в {args.file}")

if __name__ == '__main__':
    main()
