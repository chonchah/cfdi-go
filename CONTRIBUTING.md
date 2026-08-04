# Contributing

¡Gracias por tu interés en contribuir! Estas son las pautas para hacerlo de forma sencilla y consistente.

## Cómo contribuir

1. **Reporta un issue** si encuentras un bug o quieres proponer una mejora.
2. **Crea un fork** del repositorio y una rama propia:

```bash
git checkout -b feature/descripcion-corta
```

3. **Haz cambios** siguiendo las convenciones del proyecto.
4. **Agrega/actualiza tests** y verifica que todo pase.
5. **Envía un Pull Request** a la rama `main`.

## Convenciones

- Usa comentarios y docstrings en español o inglés, consistentes con el archivo que edites.
- Formatea el código Go con `gofmt` y el código Python siguiendo PEP 8.
- No agregues dependencias nuevas sin justificación.
- Nunca incluyas credenciales, claves o datos sensibles en el código.

## Verificación previa

### Go

```bash
go build ./...
go test ./... -v
go vet ./...
```

### Python

```bash
cd python_files
python test_cfdi.py
```

## Flujo de un Pull Request

- Describe claramente el cambio y su motivación.
- Relaciona el PR con su issue correspondiente (si existe).
- Mantén el historial limpio; usa mensajes de commit descriptivos.

Ante dudas, abre un issue antes de empezar. ¡Apreciamos tu aporte!