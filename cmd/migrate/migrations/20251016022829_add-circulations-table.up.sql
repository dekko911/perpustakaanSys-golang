CREATE TABLE
    IF NOT EXISTS `circulations` (
        `id` CHAR(36) NOT NULL,
        `id_skl` CHAR(36) NOT NULL,
        `buku_id` CHAR(36) NOT NULL,
        `peminjam` VARCHAR(100) NOT NULL,
        `tanggal_pinjam` DATE NOT NULL,
        `jatuh_tempo` DATE NOT NULL,
        `denda` DECIMAL(11, 2) DEFAULT 0.00 NOT NULL,
        `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`id`),
        UNIQUE KEY (`id_skl`),
        CONSTRAINT `fk_circulations_buku_id` FOREIGN KEY (`buku_id`) REFERENCES books (`id`) ON DELETE CASCADE ON UPDATE CASCADE
    );