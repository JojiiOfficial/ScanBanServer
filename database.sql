SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;


CREATE TABLE `BlockedIP` (
  `pk_id` int(10) UNSIGNED NOT NULL,
  `ip` varchar(16) NOT NULL,
  `reportCount` int(10) UNSIGNED NOT NULL DEFAULT '1',
  `isProxy` tinyint(1) NOT NULL DEFAULT '0',
  `validated` tinyint(1) NOT NULL DEFAULT '0',
  `lastReport` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `firstReport` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `domain` text,
  `Hostname` text,
  `type` int(10) UNSIGNED NOT NULL,
  `knownAbuser` tinyint(1) NOT NULL DEFAULT '0',
  `knownHacker` tinyint(1) NOT NULL DEFAULT '0',
  `deleted` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
DELIMITER $$
CREATE TRIGGER `Handle delete` BEFORE UPDATE ON `BlockedIP` FOR EACH ROW IF NEW.deleted = 1 THEN

SET NEW.reportCount = 0;
DELETE FROM	IPreason WHERE IPreason.ip=NEW.pk_id;
DELETE FROM	Reporter WHERE Reporter.ip=NEW.ip;

END IF
$$
DELIMITER ;

CREATE TABLE `IPreason` (
  `pk_id` int(10) UNSIGNED NOT NULL,
  `ip` int(10) UNSIGNED NOT NULL,
  `reason` int(10) UNSIGNED NOT NULL,
  `author` int(10) UNSIGNED NOT NULL,
  `added` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `IPtype` (
  `pk_id` int(11) UNSIGNED NOT NULL,
  `type` varchar(10) NOT NULL,
  `description` text NOT NULL,
  `temporary` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `IPwhitelist` (
  `ip` varchar(16) NOT NULL,
  `added` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Reason` (
  `pk_id` int(10) UNSIGNED NOT NULL,
  `description` text NOT NULL,
  `minRequestCount` tinyint(3) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Reporter` (
  `pk_id` int(10) UNSIGNED NOT NULL,
  `reporterID` int(10) UNSIGNED NOT NULL,
  `ip` varchar(16) NOT NULL,
  `reportDate` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `User` (
  `pk_id` int(10) UNSIGNED NOT NULL,
  `username` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `token` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `reportedIPs` int(10) UNSIGNED NOT NULL DEFAULT '0',
  `lastReport` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `createdAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `isValid` tinyint(1) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


ALTER TABLE `BlockedIP`
  ADD PRIMARY KEY (`pk_id`),
  ADD UNIQUE KEY `ip` (`ip`),
  ADD UNIQUE KEY `pk_id` (`pk_id`);

ALTER TABLE `IPreason`
  ADD PRIMARY KEY (`pk_id`);

ALTER TABLE `IPtype`
  ADD PRIMARY KEY (`pk_id`),
  ADD UNIQUE KEY `pk_id` (`pk_id`);

ALTER TABLE `IPwhitelist`
  ADD PRIMARY KEY (`ip`),
  ADD UNIQUE KEY `ip` (`ip`);

ALTER TABLE `Reason`
  ADD PRIMARY KEY (`pk_id`);

ALTER TABLE `Reporter`
  ADD PRIMARY KEY (`pk_id`);

ALTER TABLE `User`
  ADD PRIMARY KEY (`pk_id`);


ALTER TABLE `BlockedIP`
  MODIFY `pk_id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `IPreason`
  MODIFY `pk_id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `IPtype`
  MODIFY `pk_id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `Reason`
  MODIFY `pk_id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `Reporter`
  MODIFY `pk_id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `User`
  MODIFY `pk_id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
