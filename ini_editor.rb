# ini_editor.rb
# INI Color Editor на Ruby

require 'optparse'

RESET = "\033[0m"
CYAN = "\033[96m"
GREEN = "\033[92m"
YELLOW = "\033[93m"
GRAY = "\033[90m"
WHITE = "\033[37m"
BLUE = "\033[94m"
BLACK = "\033[30m"

DARK = {section: CYAN, key: GREEN, value: YELLOW, comment: GRAY, equals: WHITE, bracket: WHITE}
LIGHT = {section: BLUE, key: GREEN, value: YELLOW, comment: WHITE, equals: BLACK, bracket: BLACK}

class INIFile
  attr_reader :sections, :order

  def initialize
    @sections = {}
    @order = []
  end

  def parse(content)
    lines = content.lines
    current_section = ''
    @sections = {}
    @order = []
    lines.each do |line|
      trimmed = line.strip
      next if trimmed.empty? || trimmed.start_with?(';', '#')
      if trimmed.start_with?('[') && trimmed.end_with?(']')
        section = trimmed[1..-2].strip
        unless @sections[section]
          @sections[section] = {}
          @order << section
        end
        current_section = section
      elsif trimmed.include?('=') && !current_section.empty?
        key, value = trimmed.split('=', 2).map(&:strip)
        @sections[current_section][key] = value
      end
    end
    if @sections.empty?
      @sections[''] = {}
      @order << ''
    end
  end

  def get_value(section, key)
    @sections[section]&.[](key)
  end

  def set_value(section, key, value)
    unless @sections[section]
      @sections[section] = {}
      @order << section
    end
    @sections[section][key] = value
  end

  def delete_key(section, key)
    if @sections[section]&.key?(key)
      @sections[section].delete(key)
      true
    else
      false
    end
  end

  def delete_section(section)
    if @sections[section]
      @sections.delete(section)
      @order.delete(section)
      true
    else
      false
    end
  end

  def add_section(section)
    unless @sections[section]
      @sections[section] = {}
      @order << section
      true
    else
      false
    end
  end

  def to_string(use_color, theme)
    colors = theme == 'light' ? LIGHT : DARK
    lines = []
    @order.each do |section|
      unless section.empty?
        lines << "#{colors[:bracket]}[#{RESET}#{colors[:section]}#{section}#{RESET}#{colors[:bracket]}]#{RESET}"
      end
      @sections[section].each do |key, value|
        lines << "#{colors[:key]}#{key}#{RESET}#{colors[:equals]} = #{RESET}#{colors[:value]}#{value}#{RESET}"
      end
      lines << '' unless @sections[section].empty?
    end
    lines.join("\n")
  end

  def save(filename)
    File.write(filename, to_string(false, ''))
  end
end

options = {}
OptionParser.new do |opts|
  opts.banner = "Использование: ruby ini_editor.rb [опции] [файл.ini]"
  opts.on('-v', '--view', 'Показать содержимое') { options[:view] = true }
  opts.on('-s', '--set VAL', 'Установить ключ (section.key=value)') { |v| options[:set] = v }
  opts.on('-g', '--get KEY', 'Получить значение ключа (section.key)') { |v| options[:get] = v }
  opts.on('-d', '--delete KEY', 'Удалить ключ (section.key)') { |v| options[:delete] = v }
  opts.on('-a', '--add SECTION', 'Добавить секцию') { |v| options[:add] = v }
  opts.on('-r', '--remove SECTION', 'Удалить секцию') { |v| options[:remove] = v }
  opts.on('-o', '--output FILE', 'Сохранить в другой файл') { |v| options[:output] = v }
  opts.on('--no-color', 'Отключить цвета') { options[:no_color] = true }
  opts.on('--theme THEME', 'Цветовая тема (dark/light)') { |v| options[:theme] = v }
end.parse!

file = ARGV[0]
if file.nil?
  print "Введите путь к INI файлу: "
  file = gets.chomp.strip
  if file.empty?
    puts "Файл не указан."
    exit 1
  end
end

ini = INIFile.new
if File.exist?(file)
  content = File.read(file)
  ini.parse(content)
else
  puts "Файл #{file} не найден. Создаём новый."
end

set_val = options[:set]
get_key = options[:get]
del_key = options[:delete]
add_sec = options[:add]
rem_sec = options[:remove]
output = options[:output]
use_color = !options[:no_color]
theme = options[:theme] || 'dark'
view = options[:view]

if set_val
  eq = set_val.index('=')
  unless eq
    puts "Неверный формат --set. Используйте section.key=value"
    exit 1
  end
  section_key = set_val[0...eq]
  value = set_val[eq+1..-1]
  dot = section_key.index('.')
  section = dot ? section_key[0...dot].strip : ''
  key = dot ? section_key[dot+1..-1].strip : section_key.strip
  ini.set_value(section, key, value.strip)
  puts "Установлено: #{section}.#{key} = #{value}"
end

if get_key
  dot = get_key.index('.')
  section = dot ? get_key[0...dot].strip : ''
  key = dot ? get_key[dot+1..-1].strip : get_key.strip
  val = ini.get_value(section, key)
  if val
    puts val
  else
    puts "Ключ #{get_key} не найден."
  end
end

if del_key
  dot = del_key.index('.')
  section = dot ? del_key[0...dot].strip : ''
  key = dot ? del_key[dot+1..-1].strip : del_key.strip
  if ini.delete_key(section, key)
    puts "Ключ #{del_key} удалён."
  else
    puts "Ключ #{del_key} не найден."
  end
end

if add_sec
  if ini.add_section(add_sec.strip)
    puts "Секция #{add_sec} добавлена."
  else
    puts "Секция #{add_sec} уже существует."
  end
end

if rem_sec
  if ini.delete_section(rem_sec.strip)
    puts "Секция #{rem_sec} удалена."
  else
    puts "Секция #{rem_sec} не найдена."
  end
end

if view || (!set_val && !get_key && !del_key && !add_sec && !rem_sec)
  out = ini.to_string(use_color, theme)
  if output
    ini.save(output)
    puts "Сохранено в #{output}"
  else
    puts out
  end
end

if (set_val || del_key || add_sec || rem_sec) && !output
  ini.save(file)
  puts "Изменения сохранены в #{file}"
end
