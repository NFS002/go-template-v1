INSERT INTO users 
(first_name, last_name, email, password, scope)
VALUES ('User','One','user@example.com','$2a$12$XGCeXWIpjEZp7DU/12wct.lpQ/TA7YLvBfdBUQqcOt7Ko8AXUYSsC', 'read:a,write:a,read:b,write:b');
-- email: user@example.com
-- password: 'secret'

INSERT INTO users 
(first_name, last_name, email, password, scope)
VALUES ('User','Two','user2@example.com','$2a$12$XGCeXWIpjEZp7DU/12wct.lpQ/TA7YLvBfdBUQqcOt7Ko8AXUYSsC', 'read:a,read:b');
-- email: user2@example.com
-- password: 'secret'

INSERT INTO users 
(first_name, last_name, email, password, scope)
VALUES ('User','Three','user3@example.com','$2a$12$XGCeXWIpjEZp7DU/12wct.lpQ/TA7YLvBfdBUQqcOt7Ko8AXUYSsC', 'write:b');
-- email: user3@example.com
-- password: 'secret'