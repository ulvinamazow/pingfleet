```markdown
# URL Checker

Bu proqram verilmiş URL-ləri paralel olaraq yoxlayır, hər birinin HTTP status kodunu, cavab müddətini və mümkün xətalarını göstərir. Əlavə olaraq təkrar cəhd (retry), fayldan URL oxuma və JSON çıxış kimi funksiyalar dəstəkləyir.

## Xüsusiyyətlər

- **Paralel yoxlama**: URL-lər konfiqurasiya olunan sayda goroutine ilə eyni anda yoxlanılır.
- **Təkrar cəhd (retry)**: 5xx xətaları və ya şəbəkə problemləri zamanı avtomatik təkrar yoxlama aparılır.
- **Fayl dəstəyi**: URL siyahısını mətn faylından oxumaq mümkündür.
- **JSON çıxış**: Nəticələri strukturlaşdırılmış JSON formatında əldə etmək olar.
- **Konfiqurasiya olunan parametrlər**: işçi sayı, timeout, retry sayı, fayl yolu və çıxış formatı əmr sətri bayraqları ilə təyin edilir.
- **Ağıllı çıxış kodu**: 0 – bütün URL-lər uğurlu, 1 – ən azı bir problemli və ya xətalı URL, 2 – fayl oxunarkən xəta.

## Quraşdırma

Proqramı işə salmaq üçün [Go proqramlaşdırma dili](https://go.dev/dl/) quraşdırılmış olmalıdır.

Layihəni klonlayın və ya `main.go` faylını kompüterinizə yükləyin.

## İstifadə

Əsas icra əmri:

```bash
go run main.go [bayraqlar]
```

### Mövcud bayraqlar (flags)

| Bayraq      | Tip      | Standart dəyər | Təsvir |
|-------------|----------|----------------|--------|
| `-file`     | string   | `""`           | URL siyahısı olan faylın yolu. Boş qalarsa, standart siyahı istifadə olunur. |
| `-workers`  | int      | `3`            | Paralel işləyəcək goroutine sayı. |
| `-timeout`  | duration | `5s`           | Hər sorğu üçün maksimum gözləmə müddəti (məsələn, `10s`, `500ms`). |
| `-retries`  | int      | `1`            | Hər URL üçün maksimum cəhd sayı (≥1). |
| `-json`     | bool     | `false`        | `true` olarsa, nəticələr JSON formatında çap edilir. |

### Nümunələr

1. Standart URL siyahısını 5 işçi ilə yoxla, hər birinə 2 dəfə cəhd et:
   ```bash
   go run main.go -workers=5 -retries=2
   ```

2. URL siyahısını fayldan oxu və nəticələri JSON olaraq göstər:
   ```bash
   go run main.go -file=urls.txt -json
   ```

3. Timeout müddətini 10 saniyəyə qaldır və 4 işçi istifadə et:
   ```bash
   go run main.go -timeout=10s -workers=4
   ```

## Fayl formatı

`-file` bayrağı ilə verilən fayl hər sətirdə bir URL ehtiva etməlidir. Boş sətirlər və `#` ilə başlayan şərh sətirləri nəzərə alınmır.

Nümunə `urls.txt`:

```
# Əsas saytlar
https://go.dev
https://github.com
https://golang.org

# Test üçün
https://httpbin.org/status/404
```

## Çıxış nümunələri

Mətn formatında:

```
[OK     ] https://go.dev                                      -> 200 (321ms, attempt: 1)
[OK     ] https://github.com                                  -> 200 (412ms, attempt: 1)
[PROBLEM] https://httpbin.org/status/404                      -> 404 (187ms, attempt: 1)
[ERROR  ] https://this-domain-almost-certainly-does-not-exist-123.com -> dial tcp: lookup ... : no such host (attempt: 2)
```

JSON formatında (`-json` ilə):

```json
[
  {
    "url": "https://go.dev",
    "status": 200,
    "duration_ms": 321,
    "attempts": 1
  },
  {
    "url": "https://httpbin.org/status/404",
    "status": 404,
    "duration_ms": 187,
    "attempts": 1
  },
  {
    "url": "https://this-domain-almost-certainly-does-not-exist-123.com",
    "duration_ms": 1234,
    "attempts": 2,
    "error": "dial tcp: lookup ... : no such host"
  }
]
```

## Kod strukturu

- `Result` strukturu – URL, status, müddət, xəta və cəhd sayını saxlayır.
- `jsonResult` strukturu – JSON çıxışı üçün uyğunlaşdırılmış versiya.
- `checkURL` – tək URL-i yoxlayır.
- `checkWithRetries` – retry mexanizmini idarə edir.
- `worker` – goroutine olaraq URL-ləri emal edir.
- `readURLs` – fayldan URL siyahısını oxuyur.
- `main` – bayraqları analiz edir, işçiləri başladır, nəticələri toplayır və çap edir.
