package com.microservices.auth;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
public class AuthController {

    private final JdbcTemplate jdbcTemplate;

    @Value("${DB_PATH:/data/auth.db}")
    private String dbPath;

    public AuthController(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public record AuthRequest(String username, String password) {}

    @GetMapping({"/auth/status", "/health"})
    public Map<String, Object> status() {
        Integer credentialsCount = jdbcTemplate.queryForObject("SELECT COUNT(*) FROM auth_credentials", Integer.class);
        Integer sessionsCount = jdbcTemplate.queryForObject("SELECT COUNT(*) FROM auth_sessions", Integer.class);
        return Map.of(
            "service", "auth-api",
            "language", "Java / Spring Boot",
            "status", "UP",
            "database", dbPath,
            "credentials_count", credentialsCount != null ? credentialsCount : 0,
            "sessions_logged", sessionsCount != null ? sessionsCount : 0
        );
    }

    @PostMapping("/auth/login")
    public ResponseEntity<Map<String, Object>> login(@RequestBody AuthRequest req) {
        if (req.username() == null || req.username().isBlank() || req.password() == null || req.password().isBlank()) {
            return ResponseEntity.badRequest().body(Map.of(
                "valid", false,
                "message", "Debe proporcionar username y password"
            ));
        }

        String normalizedUser = req.username().trim().toLowerCase();
        String expectedHash = DatabaseInitializer.hashPassword(req.password());

        List<Map<String, Object>> rows = jdbcTemplate.queryForList(
            "SELECT id, username, password_hash, active FROM auth_credentials WHERE username = ?",
            normalizedUser
        );

        Object activeObj = !rows.isEmpty() ? rows.get(0).get("active") : null;
        boolean isActive = activeObj instanceof Number n && n.intValue() == 1;
        boolean valid = isActive
            && expectedHash.equals(rows.get(0).get("password_hash"));

        if (valid) {
            String token = "sess_" + UUID.randomUUID().toString().replace("-", "");
            jdbcTemplate.update(
                "INSERT INTO auth_sessions (username, valid, token) VALUES (?, 1, ?)",
                normalizedUser,
                token
            );
            return ResponseEntity.ok(Map.of(
                "valid", true,
                "message", "Credenciales válidas",
                "username", normalizedUser,
                "token", token
            ));
        }

        jdbcTemplate.update(
            "INSERT INTO auth_sessions (username, valid, token) VALUES (?, 0, NULL)",
            normalizedUser
        );
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of(
            "valid", false,
            "message", "Credenciales inválidas",
            "username", normalizedUser
        ));
    }

    @PostMapping("/auth/register")
    public ResponseEntity<Map<String, Object>> register(@RequestBody AuthRequest req) {
        if (req.username() == null || req.username().isBlank() || req.password() == null || req.password().length() < 4) {
            return ResponseEntity.badRequest().body(Map.of(
                "message", "Username y password (mín. 4 caracteres) requeridos"
            ));
        }
        String normalizedUser = req.username().trim().toLowerCase();
        try {
            jdbcTemplate.update(
                "INSERT INTO auth_credentials (username, password_hash) VALUES (?, ?)",
                normalizedUser,
                DatabaseInitializer.hashPassword(req.password())
            );
            return ResponseEntity.status(HttpStatus.CREATED).body(Map.of(
                "message", "Credencial registrada en auth.db",
                "username", normalizedUser
            ));
        } catch (Exception ex) {
            return ResponseEntity.status(HttpStatus.CONFLICT).body(Map.of(
                "message", "La credencial ya existe en auth.db"
            ));
        }
    }

    @GetMapping("/auth/sessions")
    public Map<String, Object> sessions() {
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(
            "SELECT id, username, valid, token, created_at FROM auth_sessions ORDER BY id DESC"
        );
        return Map.of("total", rows.size(), "data", rows);
    }
}
