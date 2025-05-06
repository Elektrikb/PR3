# PR3

## Как использовать программу
1) Сервер:
  1.Запустите серверное приложение первым
  2.Сервер создаст папку ReceivedFiles для хранения полученных файлов
  3.Файл analysis_results.txt будет содержать все результаты анализа

2) Клиент:
  1.Запустите клиентское приложение
  2.Введите путь к текстовому файлу для анализа
  3.Получите результаты анализа от сервера

## Серверная часть
```
using System;
using System.IO;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;

namespace TextAnalysisServer
{
    class Program
    {
        private const int Port = 8888;
        private static readonly string AnalysisResultsPath = Path.Combine(Directory.GetCurrentDirectory(), "analysis_results.txt");
        private static readonly string ReceivedFilesPath = Path.Combine(Directory.GetCurrentDirectory(), "ReceivedFiles");
        private static readonly object fileLock = new object();

        static async Task Main(string[] args)
        {
            // Создаем директорию для полученных файлов
            Directory.CreateDirectory(ReceivedFilesPath);

            // Инициализируем файл результатов
            InitializeResultsFile();

            var listener = new TcpListener(IPAddress.Any, Port);
            listener.Start();
            Console.WriteLine($"Сервер запущен на порту {Port}. Ожидание подключений...");

            while (true)
            {
                var client = await listener.AcceptTcpClientAsync();
                _ = HandleClientAsync(client); // Обрабатываем клиента в отдельном потоке
            }
        }

        private static void InitializeResultsFile()
        {
            if (!File.Exists(AnalysisResultsPath))
            {
                File.WriteAllText(AnalysisResultsPath, "Результаты анализа файлов:\n\n");
            }
        }

        private static async Task HandleClientAsync(TcpClient client)
        {
            try
            {
                using (client)
                using (var stream = client.GetStream())
                using (var reader = new StreamReader(stream, Encoding.UTF8))
                using (var writer = new StreamWriter(stream, Encoding.UTF8) { AutoFlush = true })
                {
                    // Читаем имя файла
                    var fileName = await reader.ReadLineAsync();
                    if (string.IsNullOrEmpty(fileName))
                    {
                        await writer.WriteLineAsync("Ошибка: Не получено имя файла");
                        return;
                    }

                    // Читаем содержимое файла
                    var fileContent = await reader.ReadToEndAsync();

                    // Генерируем уникальное имя файла
                    var uniqueFileName = GetUniqueFileName(fileName);
                    var filePath = Path.Combine(ReceivedFilesPath, uniqueFileName);

                    // Сохраняем файл
                    await File.WriteAllTextAsync(filePath, fileContent);

                    // Анализируем файл
                    var analysisResult = AnalyzeFile(filePath, fileContent);

                    // Сохраняем результаты
                    SaveAnalysisResult(uniqueFileName, analysisResult);

                    // Отправляем результаты клиенту
                    await SendAnalysisResult(writer, uniqueFileName, analysisResult);

                    Console.WriteLine($"Файл {uniqueFileName} обработан успешно");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Ошибка при обработке клиента: {ex.Message}");
            }
        }

        private static string GetUniqueFileName(string fileName)
        {
            var name = Path.GetFileNameWithoutExtension(fileName);
            var ext = Path.GetExtension(fileName);
            var timestamp = DateTime.Now.ToString("yyyyMMddHHmmssfff");
            return $"{name}_{timestamp}{ext}";
        }

        private static (int Lines, int Words, int Chars) AnalyzeFile(string filePath, string content)
        {
            var lines = content.Split('\n').Length;
            var words = content.Split(new[] { ' ', '\t', '\n', '\r' }, StringSplitOptions.RemoveEmptyEntries).Length;
            var chars = content.Length;

            return (lines, words, chars);
        }

        private static void SaveAnalysisResult(string fileName, (int Lines, int Words, int Chars) result)
        {
            lock (fileLock)
            {
                File.AppendAllText(AnalysisResultsPath, 
                    $"Файл: {fileName}\n" +
                    $"Строк: {result.Lines}, Слов: {result.Words}, Символов: {result.Chars}\n\n");
            }
        }

        private static async Task SendAnalysisResult(StreamWriter writer, string fileName, (int Lines, int Words, int Chars) result)
        {
            await writer.WriteLineAsync($"Имя файла: {fileName}");
            await writer.WriteLineAsync($"Строк: {result.Lines}, Слов: {result.Words}, Символов: {result.Chars}");
        }
    }
}
```
## Клиентская часть

```
using System;
using System.IO;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;

namespace TextAnalysisClient
{
    class Program
    {
        private const string ServerAddress = "127.0.0.1";
        private const int Port = 8888;

        static async Task Main(string[] args)
        {
            Console.WriteLine("Клиент анализа текстовых файлов");
            Console.WriteLine("Введите путь к файлу для анализа:");

            while (true)
            {
                var filePath = Console.ReadLine();

                if (string.IsNullOrWhiteSpace(filePath))
                {
                    Console.WriteLine("Неверный путь к файлу. Попробуйте снова:");
                    continue;
                }

                if (!File.Exists(filePath))
                {
                    Console.WriteLine("Файл не найден. Попробуйте снова:");
                    continue;
                }

                try
                {
                    await SendFileForAnalysis(filePath);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Ошибка: {ex.Message}");
                }

                Console.WriteLine("\nВведите путь к следующему файлу или нажмите Enter для выхода:");
                if (string.IsNullOrWhiteSpace(Console.ReadLine()))
                    break;
            }
        }

        private static async Task SendFileForAnalysis(string filePath)
        {
            using (var client = new TcpClient())
            {
                await client.ConnectAsync(ServerAddress, Port);
                Console.WriteLine("Подключено к серверу...");

                using (var stream = client.GetStream())
                using (var reader = new StreamReader(stream, Encoding.UTF8))
                using (var writer = new StreamWriter(stream, Encoding.UTF8) { AutoFlush = true })
                {
                    // Отправляем имя файла
                    var fileName = Path.GetFileName(filePath);
                    await writer.WriteLineAsync(fileName);

                    // Отправляем содержимое файла
                    var fileContent = await File.ReadAllTextAsync(filePath);
                    await writer.WriteAsync(fileContent);

                    Console.WriteLine("Файл отправлен. Ожидание результатов анализа...");

                    // Получаем и выводим результаты
                    Console.WriteLine("\nРезультаты анализа:");
                    Console.WriteLine(await reader.ReadLineAsync()); // Имя файла
                    Console.WriteLine(await reader.ReadLineAsync()); // Статистика
                }
            }
        }
    }
}
```
