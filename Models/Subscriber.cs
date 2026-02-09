using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace landing.Models;

[Table("subscribers")]
public class Subscriber
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [Column("email")]
    public string Email { get; set; } = string.Empty;
}
