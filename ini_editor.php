<?php
// ini_editor.php
// INI Color Editor на PHP

if (php_sapi_name() !== 'cli') {
    die("Это консольное приложение.\n");
}

// ANSI-цвета
define('RESET', "\033[0m");
define('CYAN', "\033[96m");
define('GREEN', "\033[92m");
define('YELLOW', "\033[93m");
define('GRAY', "\033[90m");
define('WHITE', "\033[37m");
define('BLUE', "\033[94m");
define('BLACK', "\033[30m");

$DARK = ['section'=>CYAN, 'key'=>GREEN, 'value'=>YELLOW, 'comment'=>GRAY, 'equals'=>WHITE, 'bracket'=>WHITE];
$LIGHT = ['section'=>BLUE, 'key'=>GREEN, 'value'=>YELLOW, 'comment'=>WHITE, 'equals'=>BLACK, 'bracket'=>BLACK];

class INIFile {
    private $sections = [];
    private $order = [];

    public function parse($content) {
        $lines = explode("\n", $content);
        $currentSection = '';
        $this->sections = [];
        $this->order = [];
        foreach ($lines as $line) {
            $trimmed = trim($line);
            if ($trimmed === '' || str_starts_with($trimmed, ';') || str_starts_with($trimmed, '#')) continue;
            if (str_starts_with($trimmed, '[') && str_ends_with($trimmed, ']')) {
                $section = trim(substr($trimmed, 1, -1));
                if (!isset($this->sections[$section])) {
                    $this->sections[$section] = [];
                    $this->order[] = $section;
                }
                $currentSection = $section;
            } elseif (str_contains($trimmed, '=') && $currentSection !== '') {
                list($key, $value) = explode('=', $trimmed, 2);
                $this->sections[$currentSection][trim($key)] = trim($value);
            }
        }
        if (empty($this->sections)) {
            $this->sections[''] = [];
            $this->order[] = '';
        }
    }

    public function getValue($section, $key) {
        return $this->sections[$section][$key] ?? null;
    }

    public function setValue($section, $key, $value) {
        if (!isset($this->sections[$section])) {
            $this->sections[$section] = [];
            $this->order[] = $section;
        }
        $this->sections[$section][$key] = $value;
    }

    public function deleteKey($section, $key) {
        if (isset($this->sections[$section]) && array_key_exists($key, $this->sections[$section])) {
            unset($this->sections[$section][$key]);
            return true;
        }
        return false;
    }

    public function deleteSection($section) {
        if (isset($this->sections[$section])) {
            unset($this->sections[$section]);
            $this->order = array_values(array_filter($this->order, fn($s) => $s !== $section));
            return true;
        }
        return false;
    }

    public function addSection($section) {
        if (!isset($this->sections[$section])) {
            $this->sections[$section] = [];
            $this->order[] = $section;
            return true;
        }
        return false;
    }

    public function toString($useColor, $theme) {
        $colors = $theme === 'light' ? $GLOBALS['LIGHT'] : $GLOBALS['DARK'];
        $lines = [];
        foreach ($this->order as $section) {
            if ($section !== '') {
                $lines[] = $colors['bracket'] . '[' . RESET . $colors['section'] . $section . RESET . $colors['bracket'] . ']' . RESET;
            }
            foreach ($this->sections[$section] as $key => $value) {
                $lines[] = $colors['key'] . $key . RESET . $colors['equals'] . ' = ' . RESET . $colors['value'] . $value . RESET;
            }
            if (!empty($this->sections[$section])) $lines[] = '';
        }
        return implode("\n", $lines);
    }

    public function save($filename) {
        file_put_contents($filename, $this->toString(false, ''));
    }
}

// Парсинг аргументов
$opts = getopt('vg:s:d:a:r:o:', ['view', 'set:', 'get:', 'delete:', 'add:', 'remove:', 'output:', 'no-color', 'theme:']);
$args = array_values(array_filter($argv, fn($a) => !str_starts_with($a, '-')));
$file = $args[1] ?? null;

if (!$file) {
    echo "Введите путь к INI файлу: ";
    $file = trim(fgets(STDIN));
    if (!$file) { echo "Файл не указан.\n"; exit(1); }
}

$ini = new INIFile();
if (file_exists($file)) {
    $content = file_get_contents($file);
    $ini->parse($content);
} else {
    echo "Файл $file не найден. Создаём новый.\n";
}

$setVal = $opts['set'] ?? $opts['s'] ?? null;
$getKey = $opts['get'] ?? $opts['g'] ?? null;
$delKey = $opts['delete'] ?? $opts['d'] ?? null;
$addSec = $opts['add'] ?? $opts['a'] ?? null;
$remSec = $opts['remove'] ?? $opts['r'] ?? null;
$output = $opts['output'] ?? $opts['o'] ?? null;
$noColor = isset($opts['no-color']);
$theme = $opts['theme'] ?? 'dark';
$view = isset($opts['view']) || isset($opts['v']);

if ($setVal) {
    $eq = strpos($setVal, '=');
    if ($eq === false) { echo "Неверный формат --set. Используйте section.key=value\n"; exit(1); }
    $sectionKey = substr($setVal, 0, $eq);
    $value = substr($setVal, $eq+1);
    $dot = strpos($sectionKey, '.');
    $section = $dot !== false ? substr($sectionKey, 0, $dot) : '';
    $key = $dot !== false ? substr($sectionKey, $dot+1) : $sectionKey;
    $ini->setValue(trim($section), trim($key), trim($value));
    echo "Установлено: $section.$key = $value\n";
}

if ($getKey) {
    $dot = strpos($getKey, '.');
    $section = $dot !== false ? substr($getKey, 0, $dot) : '';
    $key = $dot !== false ? substr($getKey, $dot+1) : $getKey;
    $val = $ini->getValue(trim($section), trim($key));
    if ($val !== null) echo $val . "\n";
    else echo "Ключ $getKey не найден.\n";
}

if ($delKey) {
    $dot = strpos($delKey, '.');
    $section = $dot !== false ? substr($delKey, 0, $dot) : '';
    $key = $dot !== false ? substr($delKey, $dot+1) : $delKey;
    if ($ini->deleteKey(trim($section), trim($key))) echo "Ключ $delKey удалён.\n";
    else echo "Ключ $delKey не найден.\n";
}

if ($addSec) {
    if ($ini->addSection(trim($addSec))) echo "Секция $addSec добавлена.\n";
    else echo "Секция $addSec уже существует.\n";
}

if ($remSec) {
    if ($ini->deleteSection(trim($remSec))) echo "Секция $remSec удалена.\n";
    else echo "Секция $remSec не найдена.\n";
}

if ($view || (!$setVal && !$getKey && !$delKey && !$addSec && !$remSec)) {
    $out = $ini->toString(!$noColor, $theme);
    if ($output) {
        $ini->save($output);
        echo "Сохранено в $output\n";
    } else {
        echo $out . "\n";
    }
}

if (($setVal || $delKey || $addSec || $remSec) && !$output) {
    $ini->save($file);
    echo "Изменения сохранены в $file\n";
}
