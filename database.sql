SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;


CREATE TABLE BlockedIP (
  pk_id int(10) UNSIGNED NOT NULL,
  ip varchar(16) NOT NULL,
  reportCount int(10) UNSIGNED NOT NULL DEFAULT '1',
  isProxy tinyint(1) NOT NULL,
  validated tinyint(1) NOT NULL,
  firstReport bigint(20) UNSIGNED NOT NULL,
  lastReport bigint(20) UNSIGNED NOT NULL,
  domain text,
  Hostname text,
  type int(10) UNSIGNED NOT NULL,
  knownAbuser tinyint(1) NOT NULL DEFAULT '0',
  knownHacker tinyint(1) NOT NULL DEFAULT '0',
  deleted tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
DELIMITER $$
CREATE TRIGGER `Handle update` BEFORE UPDATE ON `BlockedIP` FOR EACH ROW IF (NEW.deleted = 1) THEN

    SET NEW.reportCount = 0;
    SET NEW.lastReport=(SELECT UNIX_TIMESTAMP());
    
    DELETE FROM ReportPorts WHERE ReportPorts.reportID=(SELECT pk_id FROM Report WHERE Report.ip=NEW.pk_id);
    DELETE FROM Report WHERE Report.ip = NEW.pk_id;
    INSERT INTO FilterDelete (ip,tokenID) (SELECT DISTINCT NEW.pk_id,Token.pk_id FROM Token JOIN FilterIP ON FilterIP.filterID = Token.filter WHERE FilterIP.ip=NEW.pk_id);
    DELETE FROM FilterIP WHERE FilterIP.ip = NEW.pk_id;
    DELETE FROM IPports WHERE IPports.ip = NEW.pk_id;
    
ELSEIF (NEW.reportCount>1) THEN

SET NEW.lastReport=(SELECT UNIX_TIMESTAMP());

END IF
$$
DELIMITER ;

CREATE TABLE Filter (
  pk_id int(10) UNSIGNED NOT NULL,
  `key` tinytext,
  creationDate timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE FilterChange (
  del tinyint(1) NOT NULL COMMENT '1=delete',
  filterID int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE FilterDelete (
  pk_id int(10) UNSIGNED NOT NULL,
  ip int(10) UNSIGNED NOT NULL,
  tokenID int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE FilterIP (
  pk_id int(10) UNSIGNED NOT NULL,
  ip int(11) UNSIGNED NOT NULL,
  filterID int(11) UNSIGNED NOT NULL,
  added bigint(20) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
DELIMITER $$
CREATE TRIGGER `Remove FilterDelete if IP was inserted` AFTER INSERT ON `FilterIP` FOR EACH ROW DELETE FilterDelete FROM FilterDelete
JOIN Token ON Token.pk_id = FilterDelete.tokenID
WHERE FilterDelete.ip=NEW.ip AND Token.filter = NEW.filterID
$$
DELIMITER ;

CREATE TABLE FilterPart (
  pk_id int(10) UNSIGNED NOT NULL,
  dest tinyint(3) UNSIGNED NOT NULL,
  operator tinyint(3) UNSIGNED NOT NULL,
  val text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE FilterRow (
  pk_id int(10) UNSIGNED NOT NULL,
  filterID int(10) UNSIGNED NOT NULL,
  rowNumber tinyint(3) UNSIGNED NOT NULL,
  partID int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE GraphCache (
  graphID int(10) UNSIGNED NOT NULL,
  time int(10) UNSIGNED NOT NULL,
  value1 int(10) UNSIGNED NOT NULL,
  value2 int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IPports (
  pk_id int(10) UNSIGNED NOT NULL,
  ip int(10) UNSIGNED NOT NULL,
  port smallint(5) UNSIGNED NOT NULL,
  count int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
  added datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE KnownHostname (
  pk_id int(11) NOT NULL,
  keyword text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE PublicFilter (
  filterID int(11) NOT NULL,
  name text NOT NULL,
  description text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE Report (
  pk_id int(11) NOT NULL,
  ip int(10) UNSIGNED DEFAULT NULL,
  reporterID int(11) UNSIGNED DEFAULT NULL,
  firstReport int(10) UNSIGNED NOT NULL,
  lastReport int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ReportPorts (
  pk_id int(10) UNSIGNED NOT NULL,
  reportID int(10) UNSIGNED DEFAULT NULL,
  port smallint(5) UNSIGNED DEFAULT NULL,
  count int(10) UNSIGNED DEFAULT NULL,
  scanDate int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE Token (
  pk_id int(10) UNSIGNED NOT NULL,
  machineName text COLLATE utf8mb4_unicode_ci NOT NULL,
  token varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  filter int(10) UNSIGNED DEFAULT NULL,
  reportedIPs int(10) UNSIGNED NOT NULL DEFAULT '0',
  requests int(10) UNSIGNED NOT NULL DEFAULT '0',
  permissions tinyint(3) UNSIGNED NOT NULL DEFAULT '2',
  lastReport datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  isValid tinyint(1) NOT NULL DEFAULT '1',
  createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `User` (
  pk_id int(10) UNSIGNED NOT NULL,
  email text NOT NULL,
  username text NOT NULL,
  password text NOT NULL,
  isAdmin tinyint(1) NOT NULL DEFAULT '0',
  valid tinyint(1) NOT NULL DEFAULT '0',
  logout tinyint(1) NOT NULL DEFAULT '0',
  lastLogin timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE UserMachines (
  pk_id int(10) UNSIGNED NOT NULL,
  userID int(10) UNSIGNED NOT NULL,
  token int(10) UNSIGNED NOT NULL,
  created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


ALTER TABLE BlockedIP
  ADD PRIMARY KEY (pk_id),
  ADD UNIQUE KEY ip (ip),
  ADD UNIQUE KEY pk_id (pk_id),
  ADD KEY type (type);

ALTER TABLE `Filter`
  ADD PRIMARY KEY (pk_id),
  ADD KEY pk_id (pk_id);

ALTER TABLE FilterChange
  ADD KEY filterID (filterID);

ALTER TABLE FilterDelete
  ADD PRIMARY KEY (pk_id),
  ADD KEY ip (ip);

ALTER TABLE FilterIP
  ADD PRIMARY KEY (pk_id),
  ADD KEY ip (ip),
  ADD KEY filterID (filterID);

ALTER TABLE FilterPart
  ADD PRIMARY KEY (pk_id),
  ADD KEY pk_id (pk_id);

ALTER TABLE FilterRow
  ADD PRIMARY KEY (pk_id),
  ADD KEY partID (partID),
  ADD KEY filterID (filterID);

ALTER TABLE IPports
  ADD PRIMARY KEY (pk_id),
  ADD KEY pk_id (pk_id),
  ADD KEY ip (ip);

ALTER TABLE IPtype
  ADD PRIMARY KEY (pk_id),
  ADD UNIQUE KEY pk_id (pk_id);

ALTER TABLE IPwhitelist
  ADD PRIMARY KEY (ip),
  ADD UNIQUE KEY ip (ip);

ALTER TABLE KnownHostname
  ADD PRIMARY KEY (pk_id);

ALTER TABLE PublicFilter
  ADD PRIMARY KEY (filterID),
  ADD UNIQUE KEY filterID (filterID),
  ADD KEY filterID_2 (filterID);

ALTER TABLE Report
  ADD PRIMARY KEY (pk_id),
  ADD KEY ip (ip),
  ADD KEY reporterID (reporterID);

ALTER TABLE ReportPorts
  ADD PRIMARY KEY (pk_id),
  ADD KEY reportID (reportID);

ALTER TABLE Token
  ADD PRIMARY KEY (pk_id),
  ADD KEY filter (filter);

ALTER TABLE `User`
  ADD PRIMARY KEY (pk_id);

ALTER TABLE UserMachines
  ADD PRIMARY KEY (pk_id),
  ADD KEY userID (userID),
  ADD KEY token (token);


ALTER TABLE BlockedIP
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `Filter`
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE FilterDelete
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE FilterIP
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE FilterPart
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE FilterRow
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE IPports
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE IPtype
  MODIFY pk_id int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE KnownHostname
  MODIFY pk_id int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE Report
  MODIFY pk_id int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE ReportPorts
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE Token
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE `User`
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;
ALTER TABLE UserMachines
  MODIFY pk_id int(10) UNSIGNED NOT NULL AUTO_INCREMENT;

ALTER TABLE BlockedIP
  ADD CONSTRAINT BlockedIP_ibfk_1 FOREIGN KEY (type) REFERENCES IPtype (pk_id);

ALTER TABLE FilterDelete
  ADD CONSTRAINT FilterDelete_ibfk_1 FOREIGN KEY (ip) REFERENCES BlockedIP (pk_id);

ALTER TABLE FilterIP
  ADD CONSTRAINT FilterIP_ibfk_1 FOREIGN KEY (ip) REFERENCES BlockedIP (pk_id),
  ADD CONSTRAINT FilterIP_ibfk_2 FOREIGN KEY (filterID) REFERENCES Filter (pk_id);

ALTER TABLE FilterRow
  ADD CONSTRAINT FilterRow_ibfk_1 FOREIGN KEY (filterID) REFERENCES Filter (pk_id),
  ADD CONSTRAINT FilterRow_ibfk_2 FOREIGN KEY (partID) REFERENCES FilterPart (pk_id);

ALTER TABLE IPports
  ADD CONSTRAINT IPports_ibfk_1 FOREIGN KEY (ip) REFERENCES BlockedIP (pk_id);

ALTER TABLE Report
  ADD CONSTRAINT Report_ibfk_1 FOREIGN KEY (ip) REFERENCES BlockedIP (pk_id),
  ADD CONSTRAINT Report_ibfk_2 FOREIGN KEY (reporterID) REFERENCES Token (pk_id);

ALTER TABLE Token
  ADD CONSTRAINT Token_ibfk_1 FOREIGN KEY (filter) REFERENCES Filter (pk_id);

ALTER TABLE UserMachines
  ADD CONSTRAINT UserMachines_ibfk_1 FOREIGN KEY (userID) REFERENCES `User` (pk_id),
  ADD CONSTRAINT UserMachines_ibfk_2 FOREIGN KEY (token) REFERENCES Token (pk_id);

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
