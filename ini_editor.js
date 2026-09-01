// ini_editor.js
// INI Color Editor на JavaScript (Node.js)

const fs = require('fs');
const readline = require('readline');

// Цвета (тёмная тема)
const COLORS = {
    reset: '\x1b[0m',
    section: '\x1b[96m',
    key: '\x1b[92m',
    value: '\x1b[93m',
    comment: '\x1b[90m',
    equals: '\x1b[37m',
    bracket: '\x1b[37m',
};

const LIGHT_COLORS = {
    reset: '\x1b[0m',
    section: '\x1b[94m',
    key: '\x1b[32m',
    value: '\x1b[33m',
    comment: '\x1b[37m',
    equals: '\x1b[30m',
    bracket: '\x1b[30m',
};

class INIFile {
    constructor() {
        this.sections = {};   // section -> { key: value }
        this.order = [];      // порядок секций для сохранения
    }

    parse(content) {
        const lines = content.split(/\r?\n/);
        let currentSection = '';
        this.sections = {};
        this.order = [];
        for (let line of lines) {
            line = line.trim();
            if (!line || line.startsWith(';') || line.startsWith('#')) continue;
            if (line.startsWith('[') && line.endsWith(']')) {
                currentSection = line.slice(1, -1).trim();
                if (!this.sections[currentSection]) {
                    this.sections[currentSection] = {};
                    this.order.push(currentSection);
                }
            } else if (line.includes('=') && currentSection) {
                const [key, ...rest] = line.split('=');
                const value = rest.join('=').trim();
                this.sections[currentSection][key.trim()] = value;
            }
        }
        if (Object.keys(this.sections).length === 0) {
            this.sections[''] = {};
            this.order.push('');
        }
        return this;
    }

    getValue(section, key) {
        return this.sections[section]?.[key];
    }

    setValue(section, key, value) {
        if (!this.sections[section]) {
            this.sections[section] = {};
            this.order.push(section);
        }
        this.sections[section][key] = value;
    }

    deleteKey(section, key) {
        if (this.sections[section] && this.sections[section][key] !== undefined) {
            delete this.sections[section][key];
            return true;
        }
        return false;
    }

    deleteSection(section) {
        if (this.sections[section]) {
            delete this.sections[section];
            this.order = this.order.filter(s => s !== section);
            return true;
        }
        return false;
    }

    addSection(section) {
        if (!this.sections[section]) {
            this.sections[section] = {};
            this.order.push(section);
            return true;
        }
        return false;
    }

    toString(useColor = true, theme = 'dark') {
        const colors = theme === 'light' ? LIGHT_COLORS : COLORS;
        const lines = [];
        for (const section of this.order) {
            if (section) {
                lines.push(`${colors.bracket}[${colors.reset}${colors.section}${section}${colors.reset}${colors.bracket}]${colors.reset}`);
            }
            const keys = this.sections[section] || {};
            for (const [key, value] of Object.entries(keys)) {
                lines.push(`${colors.key}${key}${colors.reset}${colors.equals} = ${colors.reset}${colors.value}${value}${colors.reset}`);
            }
            if (Object.keys(keys).length > 0) lines.push('');
        }
        return lines.join('\n');
    }

    save(filename) {
        fs.writeFileSync(filename, this.toString(false), 'utf-8');
    }
}

function main() {
    const args = process.argv.slice(2);
    let file = null;
    let view = false;
    let setKey = null;
    let getKey = null;
    let deleteKey = null;
    let addSection = null;
    let removeSection = null;
    let outputFile = null;
    let useColor = true;
    let theme = 'dark';

    for (let i = 0; i < args.length; i++) {
        switch (args[i]) {
            case '-v': case '--view': view = true; break;
            case '-s': case '--set': setKey = args[++i]; break;
            case '-g': case '--get': getKey = args[++i]; break;
            case '-d': case '--delete': deleteKey = args[++i]; break;
            case '-a': case '--add': addSection = args[++i]; break;
            case '-r': case '--remove': removeSection = args[++i]; break;
            case '-o': case '--output': outputFile = args[++i]; break;
            case '--no-color': useColor = false; break;
            case '--color': useColor = true; break;
            case '--theme': theme = args[++i]; break;
            default:
                if (!file && !args[i].startsWith('-')) file = args[i];
                break;
        }
    }

    if (!file) {
        const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
        rl.question('Введите путь к INI файлу: ', (answer) => {
            rl.close();
            file = answer.trim();
            if (!file) { console.log('Файл не указан.'); process.exit(1); }
            run(file);
        });
    } else {
        run(file);
    }

    function run(filename) {
        let content = '';
        let ini;
        try {
            content = fs.readFileSync(filename, 'utf-8');
            ini = new INIFile().parse(content);
        } catch (e) {
            console.log(`Файл ${filename} не найден. Создаём новый.`);
            ini = new INIFile();
            ini.sections = {};
            ini.order = [];
        }

        // Обработка команд
        if (setKey) {
            const eq = setKey.indexOf('=');
            if (eq === -1) { console.error('Неверный формат --set. Используйте section.key=value'); process.exit(1); }
            const sectionKey = setKey.slice(0, eq);
            const value = setKey.slice(eq + 1);
            const dot = sectionKey.indexOf('.');
            const section = dot !== -1 ? sectionKey.slice(0, dot).trim() : '';
            const key = dot !== -1 ? sectionKey.slice(dot + 1).trim() : sectionKey.trim();
            ini.setValue(section, key, value);
            console.log(`Установлено: ${section}.${key} = ${value}`);
        }

        if (getKey) {
            const dot = getKey.indexOf('.');
            const section = dot !== -1 ? getKey.slice(0, dot).trim() : '';
            const key = dot !== -1 ? getKey.slice(dot + 1).trim() : getKey.trim();
            const value = ini.getValue(section, key);
            if (value !== undefined) console.log(value);
            else console.log(`Ключ ${getKey} не найден.`);
        }

        if (deleteKey) {
            const dot = deleteKey.indexOf('.');
            const section = dot !== -1 ? deleteKey.slice(0, dot).trim() : '';
            const key = dot !== -1 ? deleteKey.slice(dot + 1).trim() : deleteKey.trim();
            if (ini.deleteKey(section, key)) console.log(`Ключ ${deleteKey} удалён.`);
            else console.log(`Ключ ${deleteKey} не найден.`);
        }

        if (addSection) {
            if (ini.addSection(addSection.trim())) console.log(`Секция ${addSection} добавлена.`);
            else console.log(`Секция ${addSection} уже существует.`);
        }

        if (removeSection) {
            if (ini.deleteSection(removeSection.trim())) console.log(`Секция ${removeSection} удалена.`);
            else console.log(`Секция ${removeSection} не найдена.`);
        }

        // Вывод
        if (view || (!setKey && !getKey && !deleteKey && !addSection && !removeSection)) {
            const output = ini.toString(useColor, theme);
            if (outputFile) {
                fs.writeFileSync(outputFile, ini.toString(false), 'utf-8');
                console.log(`Сохранено в ${outputFile}`);
            } else {
                console.log(output);
            }
        }

        // Сохранение изменений
        if ((setKey || deleteKey || addSection || removeSection) && !outputFile) {
            ini.save(filename);
            console.log(`Изменения сохранены в ${filename}`);
        }
    }
}

if (require.main === module) {
    main();
}
