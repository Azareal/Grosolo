CREATE TABLE `activity_stream`(
	`asid` int not null AUTO_INCREMENT,
	`actor` int not null,
	`targetUser` int not null,
	`event` varchar(50) not null,
	`elementType` varchar(50) not null,
	`elementID` int not null,
	`createdAt` datetime not null,
	`extra` varchar(200) DEFAULT '' not null,
	primary key(`asid`)
);