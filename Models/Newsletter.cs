using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace landing.Models;

[Table("newsletters")]
public class Newsletter
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [Column("subject")]
    public string Subject { get; set; } = string.Empty;

    [Required]
    [Column("body")]
    public string Body { get; set; } = string.Empty;

    [Column("sent_at")]
    public DateTime SentAt { get; set; } = DateTime.UtcNow;
}
