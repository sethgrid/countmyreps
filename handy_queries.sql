-- number of participants, where subquery shows who did the most reps
select count(*) from (select email, sum(reps.count) as total_reps from reps join user on user_id=user.id where reps.created_at > '2017-10-31' and reps.created_at < '2017-12-01' group by user_id order by total_reps desc) as foo;
