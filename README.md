# pingfleet

Paralel URL sağlamlıq yoxlayıcısı — yalnız Go standart kitabxanası ilə, worker pool və retry məntiqi ilə.

`pingfleet` bir URL siyahısını (əl ilə və ya fayldan) alır, onları konfiqurasiya oluna bilən sayda worker ilə paralel yoxlayır, uğursuz sorğuları avtomatik təkrarlayır və nəticələri mətn və ya JSON formatında çap edir. Sıfır xarici asılılıq — CI/CD boru xətlərində, uptime monitorinqində və ya sadəcə "bu linklər işləyirmi?" yoxlamasında birbaşa istifadə üçün nəzərdə tutulub.

## İşə salmaq

```bash
go run main.go
```

## Compile etmək

```bash
go build -o healthcheck main.go
./healthcheck
```

## Bayraqlar (flags)

| Bayraq       | Defolt | İzah                                              |
|--------------|--------|----------------------------------------------------|
| `-file`      | ""     | URL siyahısı olan fayl (hər sətirdə bir URL, `#` ilə şərh) |
| `-workers`   | 3      | Paralel worker sayı                                |
| `-timeout`   | 5s     | Ümumi timeout (məs. `-timeout=10s`)                |
| `-retries`   | 1      | Hər URL üçün maksimum cəhd sayı                    |
| `-json`      | false  | Nəticəni JSON formatında çap et                    |

Nümunə:

```bash
./healthcheck -file=urls.txt -workers=5 -retries=2 -json
```

Heç bir xarici paket tələb olunmur — yalnız Go-nun standart kitabxanası (`net/http`, `context`, `sync`, `time`, `flag`, `bufio`, `encoding/json`).
