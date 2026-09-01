// IniEditor.java
// INI Color Editor на Java

import java.io.*;
import java.nio.file.*;
import java.util.*;

public class IniEditor {
    private static final String RESET = "\u001B[0m";
    private static final String CYAN = "\u001B[96m";
    private static final String GREEN = "\u001B[92m";
    private static final String YELLOW = "\u001B[93m";
    private static final String GRAY = "\u001B[90m";
    private static final String WHITE = "\u001B[37m";
    private static final String BLUE = "\u001B[94m";
    private static final String BLACK = "\u001B[30m";

    private static class ColorTheme {
        String section, key, value, comment, equals, bracket;
        ColorTheme(String s, String k, String v, String c, String e, String b) {
            section=s; key=k; value=v; comment=c; equals=e; bracket=b;
        }
    }

    private static final ColorTheme DARK = new ColorTheme(CYAN, GREEN, YELLOW, GRAY, WHITE, WHITE);
    private static final ColorTheme LIGHT = new ColorTheme(BLUE, GREEN, YELLOW, WHITE, BLACK, BLACK);

    private Map<String, Map<String, String>> sections = new LinkedHashMap<>();
    private List<String> order = new ArrayList<>();

    public void parse(String content) {
        String[] lines = content.split("\\r?\\n");
        String currentSection = "";
        sections.clear();
        order.clear();
        for (String line : lines) {
            String trimmed = line.trim();
            if (trimmed.isEmpty() || trimmed.startsWith(";") || trimmed.startsWith("#")) continue;
            if (trimmed.startsWith("[") && trimmed.endsWith("]")) {
                String section = trimmed.substring(1, trimmed.length()-1).trim();
                if (!sections.containsKey(section)) {
                    sections.put(section, new LinkedHashMap<>());
                    order.add(section);
                }
                currentSection = section;
            } else if (trimmed.contains("=") && !currentSection.isEmpty()) {
                int eq = trimmed.indexOf('=');
                String key = trimmed.substring(0, eq).trim();
                String value = trimmed.substring(eq+1).trim();
                sections.get(currentSection).put(key, value);
            }
        }
        if (sections.isEmpty()) {
            sections.put("", new LinkedHashMap<>());
            order.add("");
        }
    }

    public String getValue(String section, String key) {
        Map<String, String> s = sections.get(section);
        return s != null ? s.get(key) : null;
    }

    public void setValue(String section, String key, String value) {
        if (!sections.containsKey(section)) {
            sections.put(section, new LinkedHashMap<>());
            order.add(section);
        }
        sections.get(section).put(key, value);
    }

    public boolean deleteKey(String section, String key) {
        Map<String, String> s = sections.get(section);
        if (s != null && s.containsKey(key)) {
            s.remove(key);
            return true;
        }
        return false;
    }

    public boolean deleteSection(String section) {
        if (sections.containsKey(section)) {
            sections.remove(section);
            order.remove(section);
            return true;
        }
        return false;
    }

    public boolean addSection(String section) {
        if (!sections.containsKey(section)) {
            sections.put(section, new LinkedHashMap<>());
            order.add(section);
            return true;
        }
        return false;
    }

    public String toString(boolean useColor, String theme) {
        ColorTheme t = theme.equals("light") ? LIGHT : DARK;
        StringBuilder sb = new StringBuilder();
        for (String section : order) {
            if (!section.isEmpty()) {
                sb.append(t.bracket).append("[").append(RESET)
                  .append(t.section).append(section).append(RESET)
                  .append(t.bracket).append("]").append(RESET).append('\n');
            }
            Map<String, String> keys = sections.get(section);
            if (keys != null) {
                for (Map.Entry<String, String> e : keys.entrySet()) {
                    sb.append(t.key).append(e.getKey()).append(RESET)
                      .append(t.equals).append(" = ").append(RESET)
                      .append(t.value).append(e.getValue()).append(RESET).append('\n');
                }
                if (!keys.isEmpty()) sb.append('\n');
            }
        }
        return sb.toString();
    }

    public void save(String filename) throws IOException {
        Files.write(Paths.get(filename), toString(false, "").getBytes());
    }

    public static void main(String[] args) throws Exception {
        String file = null;
        boolean view = false;
        String setVal = null, getKey = null, delKey = null, addSec = null, remSec = null;
        String output = null;
        boolean noColor = false;
        String theme = "dark";

        for (int i = 0; i < args.length; i++) {
            switch (args[i]) {
                case "-v": case "--view": view = true; break;
                case "-s": case "--set": setVal = args[++i]; break;
                case "-g": case "--get": getKey = args[++i]; break;
                case "-d": case "--delete": delKey = args[++i]; break;
                case "-a": case "--add": addSec = args[++i]; break;
                case "-r": case "--remove": remSec = args[++i]; break;
                case "-o": case "--output": output = args[++i]; break;
                case "--no-color": noColor = true; break;
                case "--theme": theme = args[++i]; break;
                default:
                    if (!args[i].startsWith("-")) file = args[i];
                    break;
            }
        }

        if (file == null) {
            System.out.print("Введите путь к INI файлу: ");
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
            file = reader.readLine().trim();
            if (file.isEmpty()) {
                System.out.println("Файл не указан.");
                System.exit(1);
            }
        }

        IniEditor ini = new IniEditor();
        try {
            String content = new String(Files.readAllBytes(Paths.get(file)));
            ini.parse(content);
        } catch (IOException e) {
            System.out.println("Файл не найден. Создаём новый.");
        }

        if (setVal != null) {
            int eq = setVal.indexOf('=');
            if (eq == -1) { System.err.println("Неверный формат --set. Используйте section.key=value"); System.exit(1); }
            String sectionKey = setVal.substring(0, eq);
            String value = setVal.substring(eq+1);
            int dot = sectionKey.indexOf('.');
            String section = dot != -1 ? sectionKey.substring(0, dot).trim() : "";
            String key = dot != -1 ? sectionKey.substring(dot+1).trim() : sectionKey.trim();
            ini.setValue(section, key, value);
            System.out.println("Установлено: " + section + "." + key + " = " + value);
        }

        if (getKey != null) {
            int dot = getKey.indexOf('.');
            String section = dot != -1 ? getKey.substring(0, dot).trim() : "";
            String key = dot != -1 ? getKey.substring(dot+1).trim() : getKey.trim();
            String val = ini.getValue(section, key);
            if (val != null) System.out.println(val);
            else System.out.println("Ключ " + getKey + " не найден.");
        }

        if (delKey != null) {
            int dot = delKey.indexOf('.');
            String section = dot != -1 ? delKey.substring(0, dot).trim() : "";
            String key = dot != -1 ? delKey.substring(dot+1).trim() : delKey.trim();
            if (ini.deleteKey(section, key)) System.out.println("Ключ " + delKey + " удалён.");
            else System.out.println("Ключ " + delKey + " не найден.");
        }

        if (addSec != null) {
            if (ini.addSection(addSec.trim())) System.out.println("Секция " + addSec + " добавлена.");
            else System.out.println("Секция " + addSec + " уже существует.");
        }

        if (remSec != null) {
            if (ini.deleteSection(remSec.trim())) System.out.println("Секция " + remSec + " удалена.");
            else System.out.println("Секция " + remSec + " не найдена.");
        }

        if (view || (setVal == null && getKey == null && delKey == null && addSec == null && remSec == null)) {
            String out = ini.toString(!noColor, theme);
            if (output != null) {
                ini.save(output);
                System.out.println("Сохранено в " + output);
            } else {
                System.out.println(out);
            }
        }

        if ((setVal != null || delKey != null || addSec != null || remSec != null) && output == null) {
            ini.save(file);
            System.out.println("Изменения сохранены в " + file);
        }
    }
}
