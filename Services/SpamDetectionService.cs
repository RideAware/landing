using System.Text.RegularExpressions;

namespace landing.Services;

public class SpamDetectionService
{
    private static readonly string[] SpamPatterns =
    {
        "viagra", "cialis", "casino", "lottery", "prize",
        "click here", "buy now", "limited time",
        "congratulations", "you have won", "claim your",
        "bitcoin", "crypto", "forex", "trading bot",
        "free money", "make money fast", "work from home",
        "nigerian", "inheritance", "transfer funds",
        "<!--", "javascript:", "onclick=", "<script",
        "sveiki", "ciao", "hola", "\u043f\u0440\u0438\u0432\u0435\u0442",
        "harga", "karna", "anda", "dari",
        "toughalia", "comfythings",
        "robertgok"
    };

    private static readonly string[] EnglishWords =
    {
        "the", "and", "is", "to", "of", "for", "that", "with", "this", "have",
        "from", "be", "are", "was", "were", "been", "i", "you", "he", "she",
        "we", "they", "my", "your", "his", "her", "it", "what", "which", "who",
        "when", "where", "why", "how", "can", "will", "would", "should", "could",
        "do", "does", "did", "get", "got", "go", "going", "make", "made", "know",
        "think", "want", "need", "like", "help", "work", "use", "ask", "say", "tell",
        "give", "find", "tell", "become", "leave", "feel", "try", "ask", "need",
        "meet", "include", "continue", "set", "learn", "change", "lead", "understand"
    };

    private static readonly HashSet<string> CommonPairs = new()
    {
        "th", "he", "in", "er", "an",
        "ed", "nd", "to", "en", "ti",
        "es", "or", "te", "ar", "ou",
        "it", "ha", "is", "co", "me",
        "we", "be", "se", "as", "de",
        "so", "re", "st", "up", "at",
        "ai", "al", "il", "le", "li"
    };

    private static readonly HashSet<string> ValidSubjects = new()
    {
        "general", "support", "partnership", "feedback", "other"
    };

    private static readonly Regex UrlRegex = new(@"https?://", RegexOptions.Compiled);
    private static readonly Regex EmailRegex = new(@"[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}", RegexOptions.Compiled);
    private static readonly Regex PhoneRegex = new(@"\+?[0-9]{7,}", RegexOptions.Compiled);

    public bool IsValidName(string name)
    {
        if (name.Length < 2 || name.Length > 100)
            return false;

        int numberCount = name.Count(char.IsDigit);
        if (numberCount > 0 && (double)numberCount / name.Length > 0.33)
            return false;

        if (name.Contains("http") || name.Contains("://"))
            return false;

        return true;
    }

    public bool IsValidEmail(string email)
    {
        if (!email.Contains('@') || !email.Contains('.'))
            return false;

        var parts = email.Split('@');
        if (parts.Length != 2)
            return false;

        if (parts[0].Length < 1 || parts[0].Length > 64)
            return false;

        if (parts[1].Length < 3 || parts[1].Length > 255)
            return false;

        var domainParts = parts[1].Split('.');
        if (domainParts.Length < 2)
            return false;

        foreach (var label in domainParts)
        {
            if (label.Length < 1 || label.Length > 63)
                return false;
        }

        return true;
    }

    public bool IsValidSubject(string subject) => ValidSubjects.Contains(subject);

    public bool IsEnglishText(string text)
    {
        if (string.IsNullOrEmpty(text))
            return true;

        var lowerText = text.ToLowerInvariant();
        var words = lowerText.Split(
            new[] { ' ', '\t', '\n', '\r', '.', ',', '!', '?', ';', ':', '-', '(', ')', '[', ']', '{', '}', '"', '\'' },
            StringSplitOptions.RemoveEmptyEntries
        );

        int englishWordCount = 0;
        int totalWords = 0;

        foreach (var word in words)
        {
            if (word.Length == 0) continue;
            totalWords++;

            if (EnglishWords.Contains(word))
                englishWordCount++;
        }

        if (text.Length < 50)
            return englishWordCount >= 1;

        if (text.Length < 200)
            return englishWordCount >= 2;

        if (totalWords > 0)
            return (double)englishWordCount / totalWords >= 0.1;

        return true;
    }

    public bool IsSpamMessage(string message)
    {
        var lowerMsg = message.ToLowerInvariant();

        // Check spam keywords
        foreach (var pattern in SpamPatterns)
        {
            if (lowerMsg.Contains(pattern))
                return true;
        }

        // Check for excessive URLs
        if (UrlRegex.Matches(lowerMsg).Count > 1)
            return true;

        // Check for email addresses in message
        if (EmailRegex.IsMatch(lowerMsg))
            return true;

        // Check for phone numbers
        if (PhoneRegex.IsMatch(lowerMsg))
            return true;

        // Check for excessive exclamation marks
        if (lowerMsg.Count(c => c == '!') > 2)
            return true;

        // Check for repeated characters
        if (lowerMsg.Contains("!!!") || lowerMsg.Contains("???") || lowerMsg.Contains("..."))
            return true;

        // Check for all caps
        if (lowerMsg.Length > 20)
        {
            int letterCount = 0;
            int capsCount = 0;
            foreach (var c in message)
            {
                if (char.IsLetter(c))
                {
                    letterCount++;
                    if (char.IsUpper(c))
                        capsCount++;
                }
            }
            if (letterCount > 0 && (double)capsCount / letterCount > 0.6)
                return true;
        }

        // Check for repeated words
        var words = lowerMsg.Split(' ', StringSplitOptions.RemoveEmptyEntries);
        if (words.Length > 5)
        {
            var wordCount = new Dictionary<string, int>();
            foreach (var word in words)
            {
                wordCount.TryGetValue(word, out int count);
                wordCount[word] = count + 1;
            }
            if (wordCount.Values.Any(c => c > 3))
                return true;
        }

        // Check message length - very short messages are often spam
        if (message.Length < 15)
            return true;

        // Check for gibberish - high ratio of uncommon character transitions
        int uncommonCount = 0;
        for (int i = 0; i < lowerMsg.Length - 1; i++)
        {
            char c = lowerMsg[i];
            char next = lowerMsg[i + 1];

            if (c >= 'a' && c <= 'z' && next >= 'a' && next <= 'z')
            {
                var pair = new string(new[] { c, next });
                if (!CommonPairs.Contains(pair) && c != next)
                    uncommonCount++;
            }
        }

        if (lowerMsg.Length > 30 && uncommonCount > lowerMsg.Length / 3)
            return true;

        return false;
    }
}
