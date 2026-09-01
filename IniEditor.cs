// IniEditor.cs
// INI Color Editor на C#

using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

class IniEditor
{
    private const string RESET = "\u001B[0m";
    private const string CYAN = "\u001B[96m";
    private const string GREEN = "\u001B[92m";
    private const string YELLOW = "\u001B[93m";
    private const string GRAY = "\u001B[90m";
    private const string WHITE = "\u001B[37m";
    private const string BLUE = "\u001B[94m";
    private const string BLACK = "\u001B[30m";

    private class ColorTheme
    {
        public string Section, Key, Value, Comment, Equals, Bracket;
        public ColorTheme(string s, string k, string v, string c, string e, string b)
        { Section=s; Key=k; Value=v; Comment=c; Equals=e; Bracket=b; }
    }

    private static readonly ColorTheme Dark = new ColorTheme(CYAN, GREEN, YELLOW, GRAY, WHITE, WHITE);
    private static readonly ColorTheme Light = new ColorTheme(BLUE, GREEN, YELLOW, WHITE, BLACK, BLACK);

    private Dictionary<string, Dictionary<string, string>> sections = new Dictionary<string, Dictionary<string, string>>();
    private List<string> order = new List<string>();

    public void Parse(string content)
    {
        var lines = content.Split(new[] { "\r\n", "\n" }, StringSplitOptions.None);
        string currentSection = "";
        sections.Clear();
        order.Clear();
        foreach (var line in lines)
        {
            var trimmed = line.Trim();
            if (string.IsNullOrEmpty(trimmed) || trimmed.StartsWith(";") || trimmed.StartsWith("#"))
                continue;
            if (trimmed.StartsWith("[") && trimmed.EndsWith("]"))
            {
                var section = trimmed.Substring(1, trimmed.Length-2).Trim();
                if (!sections.ContainsKey(section))
                {
                    sections[section] = new Dictionary<string, string>();
                    order.Add(section);
                }
                currentSection = section;
            }
            else if (trimmed.Contains("=") && !string.IsNullOrEmpty(currentSection))
            {
                var eq = trimmed.IndexOf('=');
                var key = trimmed.Substring(0, eq).Trim();
                var value = trimmed.Substring(eq+1).Trim();
                sections[currentSection][key] = value;
            }
        }
        if (!sections.Any())
        {
            sections[""] = new Dictionary<string, string>();
            order.Add("");
        }
    }

    public string GetValue(string section, string key)
    {
        return sections.TryGetValue(section, out var s) ? s.GetValueOrDefault(key) : null;
    }

    public void SetValue(string section, string key, string value)
    {
        if (!sections.ContainsKey(section))
        {
            sections[section] = new Dictionary<string, string>();
            order.Add(section);
        }
        sections[section][key] = value;
    }

    public bool DeleteKey(string section, string key)
    {
        if (sections.TryGetValue(section, out var s) && s.ContainsKey(key))
        {
            s.Remove(key);
            return true;
        }
        return false;
    }

    public bool DeleteSection(string section)
    {
        if (sections.ContainsKey(section))
        {
            sections.Remove(section);
            order.Remove(section);
            return true;
        }
        return false;
    }

    public bool AddSection(string section)
    {
        if (!sections.ContainsKey(section))
        {
            sections[section] = new Dictionary<string, string>();
            order.Add(section);
            return true;
        }
        return false;
    }

    public string ToString(bool useColor, string theme)
    {
        var t = theme == "light" ? Light : Dark;
        var lines = new List<string>();
        foreach (var section in order)
        {
            if (!string.IsNullOrEmpty(section))
                lines.Add($"{t.Bracket}[{RESET}{t.Section}{section}{RESET}{t.Bracket}]{RESET}");
            var keys = sections[section];
            foreach (var kv in keys)
                lines.Add($"{t.Key}{kv.Key}{RESET}{t.Equals} = {RESET}{t.Value}{kv.Value}{RESET}");
            if (keys.Any()) lines.Add("");
        }
        return string.Join("\n", lines);
    }

    public void Save(string filename)
    {
        File.WriteAllText(filename, ToString(false, ""));
    }

    static void Main(string[] args)
    {
        string file = null;
        bool view = false;
        string setVal = null, getKey = null, delKey = null, addSec = null, remSec = null;
        string output = null;
        bool noColor = false;
        string theme = "dark";

        for (int i = 0; i < args.Length; i++)
        {
            switch (args[i])
            {
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
                    if (!args[i].StartsWith("-")) file = args[i];
                    break;
            }
        }

        if (string.IsNullOrEmpty(file))
        {
            Console.Write("Введите путь к INI файлу: ");
            file = Console.ReadLine().Trim();
            if (string.IsNullOrEmpty(file)) { Console.WriteLine("Файл не указан."); return; }
        }

        var ini = new IniEditor();
        try
        {
            var content = File.ReadAllText(file);
            ini.Parse(content);
        }
        catch
        {
            Console.WriteLine($"Файл {file} не найден. Создаём новый.");
        }

        if (setVal != null)
        {
            var eq = setVal.IndexOf('=');
            if (eq == -1) { Console.WriteLine("Неверный формат --set. Используйте section.key=value"); return; }
            var sectionKey = setVal.Substring(0, eq);
            var value = setVal.Substring(eq+1);
            var dot = sectionKey.IndexOf('.');
            var section = dot != -1 ? sectionKey.Substring(0, dot).Trim() : "";
            var key = dot != -1 ? sectionKey.Substring(dot+1).Trim() : sectionKey.Trim();
            ini.SetValue(section, key, value);
            Console.WriteLine($"Установлено: {section}.{key} = {value}");
        }

        if (getKey != null)
        {
            var dot = getKey.IndexOf('.');
            var section = dot != -1 ? getKey.Substring(0, dot).Trim() : "";
            var key = dot != -1 ? getKey.Substring(dot+1).Trim() : getKey.Trim();
            var val = ini.GetValue(section, key);
            if (val != null) Console.WriteLine(val);
            else Console.WriteLine($"Ключ {getKey} не найден.");
        }

        if (delKey != null)
        {
            var dot = delKey.IndexOf('.');
            var section = dot != -1 ? delKey.Substring(0, dot).Trim() : "";
            var key = dot != -1 ? delKey.Substring(dot+1).Trim() : delKey.Trim();
            if (ini.DeleteKey(section, key)) Console.WriteLine($"Ключ {delKey} удалён.");
            else Console.WriteLine($"Ключ {delKey} не найден.");
        }

        if (addSec != null)
        {
            if (ini.AddSection(addSec.Trim())) Console.WriteLine($"Секция {addSec} добавлена.");
            else Console.WriteLine($"Секция {addSec} уже существует.");
        }

        if (remSec != null)
        {
            if (ini.DeleteSection(remSec.Trim())) Console.WriteLine($"Секция {remSec} удалена.");
            else Console.WriteLine($"Секция {remSec} не найдена.");
        }

        if (view || (setVal == null && getKey == null && delKey == null && addSec == null && remSec == null))
        {
            var outText = ini.ToString(!noColor, theme);
            if (output != null)
            {
                ini.Save(output);
                Console.WriteLine($"Сохранено в {output}");
            }
            else
            {
                Console.WriteLine(outText);
            }
        }

        if ((setVal != null || delKey != null || addSec != null || remSec != null) && output == null)
        {
            ini.Save(file);
            Console.WriteLine($"Изменения сохранены в {file}");
        }
    }
}
