-- UserAccountType table
CREATE TABLE IF NOT EXISTS UserAccountType (
    UserRoleID BIGSERIAL PRIMARY KEY,
    RoleDesc VARCHAR(255) NOT NULL,
    Status SMALLINT DEFAULT 1
);

INSERT INTO UserAccountType (RoleDesc, Status)
SELECT 'Admin', 1
WHERE NOT EXISTS (SELECT 1 FROM UserAccountType WHERE RoleDesc = 'Admin');

-- Contacts table
CREATE TABLE IF NOT EXISTS Contacts (
    ContactID SERIAL PRIMARY KEY,
    ContactNumber VARCHAR(50) NOT NULL,
    EmailAddress VARCHAR(255),
    TelephoneNumber VARCHAR(50)
);

-- Address table
CREATE TABLE IF NOT EXISTS Address (
    AddressID SERIAL PRIMARY KEY,
    MainAddress VARCHAR(255) NOT NULL,
    SecondaryAddress VARCHAR(255),
    LastAddress VARCHAR(255)
);

-- UserInfo table
CREATE TABLE IF NOT EXISTS UserInfo (
    UserID SERIAL PRIMARY KEY,
    FullName VARCHAR(210),
    FName VARCHAR(70),
    LName VARCHAR(70),
    MName VARCHAR(70),
    StaffID VARCHAR(40),
    ContactID INT,
    AddressID INT,
    Birthdate TIMESTAMP,
    UserRoleID BIGINT,
    CONSTRAINT FK_UserInfo_UserAccountType FOREIGN KEY (UserRoleID) REFERENCES UserAccountType(UserRoleID),
    CONSTRAINT FK_UserInfo_Contacts FOREIGN KEY (ContactID) REFERENCES Contacts(ContactID),
    CONSTRAINT FK_UserInfo_Address FOREIGN KEY (AddressID) REFERENCES Address(AddressID)
);

-- UserCredentials table
CREATE TABLE IF NOT EXISTS UserCredentials (
    CredentialID SERIAL PRIMARY KEY,
    UserID INT REFERENCES UserInfo(UserID) ON DELETE CASCADE,
    Username VARCHAR(120) UNIQUE NOT NULL,
    UserPassword VARCHAR(200) NOT NULL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FailedAttempts INT DEFAULT 0,
    IsEnabled SMALLINT DEFAULT 1,
    LastPwChange TIMESTAMP,
    LastCredReset TIMESTAMP,
    IsLogin SMALLINT DEFAULT 0
);

-- AddUser function
CREATE OR REPLACE FUNCTION AddUser(
    p_fullname VARCHAR,
    p_fname VARCHAR,
    p_lname VARCHAR,
    p_mname VARCHAR DEFAULT NULL,
    p_staffid VARCHAR DEFAULT NULL,
    p_contactnumber VARCHAR DEFAULT NULL,
    p_email VARCHAR DEFAULT NULL,
    p_telephonenumber VARCHAR DEFAULT NULL,
    p_mainaddress VARCHAR DEFAULT NULL,
    p_secondaryaddress VARCHAR DEFAULT NULL,
    p_lastaddress VARCHAR DEFAULT NULL,
    p_birthdate TIMESTAMP DEFAULT NULL,
    p_userroleid BIGINT DEFAULT NULL,
    p_userpassword VARCHAR DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    new_userid INT;
    new_contactid INT;
    new_addressid INT;
    generated_username VARCHAR(120);
BEGIN
    INSERT INTO Contacts (ContactNumber, EmailAddress, TelephoneNumber)
    VALUES (p_contactnumber, p_email, p_telephonenumber)
    RETURNING ContactID INTO new_contactid;

    INSERT INTO Address (MainAddress, SecondaryAddress, LastAddress)
    VALUES (p_mainaddress, p_secondaryaddress, p_lastaddress)
    RETURNING AddressID INTO new_addressid;

    INSERT INTO UserInfo (FullName, FName, LName, MName, StaffID, ContactID, AddressID, Birthdate, UserRoleID)
    VALUES (p_fullname, p_fname, p_lname, p_mname, p_staffid, new_contactid, new_addressid, p_birthdate, p_userroleid)
    RETURNING UserID INTO new_userid;

    generated_username := COALESCE(p_fname, '') || COALESCE(p_mname, '') || COALESCE(p_lname, '');

    INSERT INTO UserCredentials (UserID, Username, UserPassword, DateCreated, FailedAttempts, IsEnabled, LastPwChange, LastCredReset, IsLogin)
    VALUES (new_userid, generated_username, p_userpassword, CURRENT_TIMESTAMP, 0, 1, NULL, NULL, 0);
END;
$$ LANGUAGE plpgsql;

-- updateuser function
CREATE OR REPLACE FUNCTION public.updateuser(
    p_userid INT,
    p_fullname VARCHAR DEFAULT NULL,
    p_fname VARCHAR DEFAULT NULL,
    p_lname VARCHAR DEFAULT NULL,
    p_mname VARCHAR DEFAULT NULL,
    p_staffid VARCHAR DEFAULT NULL,
    p_contactnumber VARCHAR DEFAULT NULL,
    p_email VARCHAR DEFAULT NULL,
    p_telephonenumber VARCHAR DEFAULT NULL,
    p_mainaddress VARCHAR DEFAULT NULL,
    p_secondaryaddress VARCHAR DEFAULT NULL,
    p_lastaddress VARCHAR DEFAULT NULL,
    p_birthdate TIMESTAMP DEFAULT NULL,
    p_userroleid BIGINT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    new_fname VARCHAR;
    new_mname VARCHAR;
    new_lname VARCHAR;
    new_username VARCHAR;
    current_contactid INT;
    current_addressid INT;
BEGIN
    SELECT ContactID, AddressID INTO current_contactid, current_addressid
    FROM UserInfo WHERE UserID = p_userid;

    UPDATE Contacts SET
        ContactNumber   = COALESCE(p_contactnumber, ContactNumber),
        EmailAddress    = COALESCE(p_email, EmailAddress),
        TelephoneNumber = COALESCE(p_telephonenumber, TelephoneNumber)
    WHERE ContactID = current_contactid;

    UPDATE Address SET
        MainAddress      = COALESCE(p_mainaddress, MainAddress),
        SecondaryAddress = COALESCE(p_secondaryaddress, SecondaryAddress),
        LastAddress      = COALESCE(p_lastaddress, LastAddress)
    WHERE AddressID = current_addressid;

    UPDATE UserInfo SET
        FullName   = COALESCE(p_fullname, FullName),
        FName      = COALESCE(p_fname, FName),
        LName      = COALESCE(p_lname, LName),
        MName      = COALESCE(p_mname, MName),
        StaffID    = COALESCE(p_staffid, StaffID),
        Birthdate  = COALESCE(p_birthdate, Birthdate),
        UserRoleID = COALESCE(p_userroleid, UserRoleID)
    WHERE UserID = p_userid;

    SELECT FName, MName, LName INTO new_fname, new_mname, new_lname
    FROM UserInfo WHERE UserID = p_userid;

    new_username := COALESCE(new_fname, '') || COALESCE(new_mname, '') || COALESCE(new_lname, '');

    UPDATE UserCredentials SET Username = new_username WHERE UserID = p_userid;
END;
$$ LANGUAGE plpgsql;

-- deleteuser function
CREATE OR REPLACE FUNCTION public.deleteuser(p_userid INT)
RETURNS VOID AS $$
DECLARE
    current_contactid INT;
    current_addressid INT;
BEGIN
    SELECT ContactID, AddressID INTO current_contactid, current_addressid
    FROM UserInfo WHERE UserID = p_userid;

    DELETE FROM UserCredentials WHERE UserID = p_userid;
    DELETE FROM UserInfo WHERE UserID = p_userid;

    IF current_contactid IS NOT NULL THEN
        DELETE FROM Contacts WHERE ContactID = current_contactid;
    END IF;

    IF current_addressid IS NOT NULL THEN
        DELETE FROM Address WHERE AddressID = current_addressid;
    END IF;
END;
$$ LANGUAGE plpgsql;
