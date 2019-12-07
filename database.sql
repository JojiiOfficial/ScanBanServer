SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;


CREATE TABLE BlockedIP (
  pk_id int(10) UNSIGNED NOT NULL,
  ip varchar(16) NOT NULL,
  reportCount int(10) UNSIGNED NOT NULL DEFAULT '1',
  isProxy tinyint(1) NOT NULL DEFAULT '0',
  validated tinyint(1) NOT NULL DEFAULT '0',
  lastReport timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  firstReport timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  domain text,
  Hostname text,
  type int(10) UNSIGNED NOT NULL,
  dyn tinyint(1) NOT NULL DEFAULT '0',
  knownAbuser tinyint(1) NOT NULL DEFAULT '0',
  knownHacker tinyint(1) NOT NULL DEFAULT '0',
  deleted tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
DELIMITER $$
CREATE TRIGGER `Handle delete` BEFORE UPDATE ON `BlockedIP` FOR EACH ROW IF (NEW.deleted = 1) THEN

SET NEW.reportCount = 0;
SET NEW.lastReport=CURRENT_TIMESTAMP;
DELETE FROM ReportPorts WHERE ReportPorts.reportID=(SELECT pk_id FROM Report WHERE Report.ip=NEW.pk_id);
DELETE FROM Report WHERE Report.ip=NEW.pk_id;


ELSEIF (NEW.reportCount>1) THEN

SET NEW.lastReport=CURRENT_TIMESTAMP;

END IF
$$
DELIMITER ;

CREATE TABLE IPtype (
  pk_id int(11) UNSIGNED NOT NULL,
  type varchar(10) NOT NULL,
  description text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO IPtype (pk_id, `type`, description) VALUES
(0, 'undefined', 'no given type'),
(1, 'hosting', 'Hosing'),
(2, 'isp', 'Internet service provider'),
(3, 'edu', 'Educational institutions'),
(4, 'gov', 'government agency'),
(5, 'mil', 'Military organization'),
(6, 'business', 'End-user organizations');

CREATE TABLE IPwhitelist (
  ip varchar(16) NOT NULL,
  added timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE KnownHostname (
  pk_id int(11) NOT NULL,
  keyword text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE Reason (
  pk_id int(10) UNSIGNED NOT NULL,
  description text NOT NULL,
  minRequestCount tinyint(3) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO Reason (pk_id, description, minRequestCount) VALUES
(1, 'Scanner', 1),
(2, 'Spammer', 5),
(3, 'Hacker', 15);

CREATE TABLE Report (
  pk_id int(11) NOT NULL,
  ip int(10) UNSIGNED DEFAULT NULL,
  reporterID int(11) UNSIGNED DEFAULT NULL,
  firstReport timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  lastReport timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ReportPorts (
  pk_id int(10) UNSIGNED NOT NULL,
  reportID int(10) UNSIGNED DEFAULT NULL,
  port int(11) DEFAULT NULL,
  count int(10) UNSIGNED DEFAULT NULL,
  scanDate int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `User` (
  pk_id int(10) UNSIGNED NOT NULL,
  username text COLLATE utf8mb4_unicode_ci NOT NULL,
  token varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  reportedIPs int(10) UNSIGNED NOT NULL DEFAULT '0',
  permissions tinyint(3) UNSIGNED NOT NULL DEFAULT '2',
  lastReport timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  isValid tinyint(1) NOT NULL DEFAULT '1',
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


ALTER TABLE BlockedIP
  ADD PRIMARY KEY (pk_id),
  ADD UNIQUE KEY ip (ip),
  ADD UNIQUE KEY pk_id (pk_id);

ALTER TABLE IPtype
  ADD PRIMARY KEY (pk_id),
  ADD UNIQUE KEY pk_id (pk_id);

ALTER TABLE IPwhitelist
  ADD PRIMARY KEY (ip),
  ADD UNIQUE KEY ip (ip);

ALTER TABLE KnownHostname
  ADD PRIMARY KEY (pk_id);

ALTER TABLE Reason
  ADD PRIMARY KEY (pk_id);

ALTER TABLE Report
  ADD PRIMARY KEY (pk_id);

ALTER TABLE ReportPorts
  ADD PRIMARY KEY (pk_id);

ALTER TABLE `User`
  ADD PRIMARY KEY (pk_id);


ALTER TABLE BlockedIP
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE IPtype
  MODIFY pk_id int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE KnownHostname
  MODIFY pk_id int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE Reason
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE Report
  MODIFY pk_id int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE ReportPorts
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `User`
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
