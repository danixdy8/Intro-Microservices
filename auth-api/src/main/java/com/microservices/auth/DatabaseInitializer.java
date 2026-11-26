package com.microservices.auth;

import jakarta.annotation.PostConstruct;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.HexFormat;

@Component
public class DatabaseInitializer {

    private final JdbcTemplate jdbcTemplate;

    public DatabaseInitializer(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public static String hashPassword(String raw) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] encoded = digest.digest(("auth_api_salt::" + raw).getBytes(StandardCharsets.UTF_8));
            return HexFormat.of().formatHex(encoded);
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException("SHA-256 no disponible", e);
        }
    }

    @PostConstruct
    public void init() {
        jdbcTemplate.execute("""
            CREATE TABLE IF NOT EXISTS auth_credentials (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL UNIQUE,
                password_hash TEXT NOT NULL,
                active INTEGER NOT NULL DEFAULT 1,
                created_at TEXT NOT NULL DEFAULT (datetime('now'))
            )
        """);

        jdbcTemplate.execute("""
            CREATE TABLE IF NOT EXISTS auth_sessions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL,
                valid INTEGER NOT NULL,
                token TEXT,
                created_at TEXT NOT NULL DEFAULT (datetime('now'))
            )
        """);

        Integer count = jdbcTemplate.queryForObject("SELECT COUNT(*) FROM auth_credentials", Integer.class);
        if (count != null && count == 0) {
            jdbcTemplate.update(
                "INSERT INTO auth_credentials (username, password_hash) VALUES (?, ?)",
                "admin@example.com",
                hashPassword("admin123")
            );
            jdbcTemplate.update(
                "INSERT INTO auth_credentials (username, password_hash) VALUES (?, ?)",
                "user@example.com",
                hashPassword("user123")
            );
        }
    }
}
