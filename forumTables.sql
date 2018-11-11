DROP TABLE IF EXISTS forum CASCADE;
DROP TABLE IF EXISTS forumUser CASCADE;
DROP TABLE IF EXISTS thread CASCADE;
DROP TABLE IF EXISTS post CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE forumUser (
    nickname citext COLLATE "C" CONSTRAINT userPK PRIMARY KEY,
    fullname VARCHAR(255) NOT NULL,
	about TEXT,
	email citext NOT NULL,
	CONSTRAINT nickname_check CHECK (nickname ~ '^[a-zA-Z0-9_.]*$'),
	CONSTRAINT unique_email UNIQUE (email)
);

CREATE TABLE forum (
    slug citext CONSTRAINT forumPK PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	userNick citext COLLATE "C",
	posts BIGINT DEFAULT 0,
	threads INT DEFAULT 0,
	CONSTRAINT forumFK FOREIGN KEY (userNick) REFERENCES forumUser (nickname)
);

CREATE TABLE thread (
	ID SERIAL CONSTRAINT threadPK PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
	author citext COLLATE "C"NOT NULL,
	forum citext,
	message TEXT NOT NULL,
	votes INT DEFAULT 0,
	slug citext,
	created timestamp,
	CONSTRAINT threadFK1 FOREIGN KEY (author) REFERENCES forumUser (nickname),
	CONSTRAINT threadFK2 FOREIGN KEY (forum) REFERENCES forum (slug)
);

CREATE TABLE post (
	ID BIGSERIAL CONSTRAINT postPK PRIMARY KEY,
	parent BIGINT,
	author citext COLLATE "C" NOT NULL,
	message TEXT NOT NULL,
	isEdited BOOLEAN DEFAULT false,
	forum citext,
	thread INT,
	created timestamp,
	path int[],
	CONSTRAINT postFK1 FOREIGN KEY (author) REFERENCES forumUser (nickname),
	CONSTRAINT postFK2 FOREIGN KEY (forum) REFERENCES forum (slug),
	CONSTRAINT postFK3 FOREIGN KEY (thread) REFERENCES thread (ID)
);


CREATE TABLE vote(
    nickname citext COLLATE "C",
	voice INT NOT NULL,
	threadId INT NOT NULL,
	CONSTRAINT voteFK1 FOREIGN KEY (nickname) REFERENCES forumUser (nickname),
	CONSTRAINT voteFK2 FOREIGN KEY (threadId) REFERENCES thread (Id),
	CONSTRAINT voiceCheck CHECK (voice = -1 OR voice = 1)	
);