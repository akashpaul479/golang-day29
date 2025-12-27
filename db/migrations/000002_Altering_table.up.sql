USE management_sys;

ALTER TABLE borrow_records 
MODIFY user_type VARCHAR(20) NOT NULL;
