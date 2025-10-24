CREATE TABLE
    IF NOT EXISTS `members` (
        `id` CHAR(36) NOT NULL,
        `id_anggota` CHAR(36) NOT NULL,
        `nama` VARCHAR(255) NOT NULL,
        `jenis_kelamin` ENUM ('L', 'P', '-') NOT NULL DEFAULT '-',
        `kelas` VARCHAR(100) NOT NULL,
        `no_telepon` VARCHAR(100) NOT NULL,
        `profil_anggota` VARCHAR(255) NULL,
        `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`id`)
    );