CREATE TABLE
    IF NOT EXISTS `books` (
        `id` CHAR(36) NOT NULL,
        `id_buku` CHAR(36) NOT NULL,
        `judul_buku` VARCHAR(255) NOT NULL,
        `cover_buku` VARCHAR(100) NULL,
        `penulis` VARCHAR(255) NOT NULL,
        `pengarang` VARCHAR(255) NOT NULL,
        `tahun` DATE NOT NULL,
        `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`id`),
        UNIQUE KEY (`id_buku`)
    );