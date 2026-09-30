-- Диалог привязан к любому объявлению пропажи/находки, не только к found.
-- Один человек может откликнуться и на «нашли», и на «ищут».

alter table dialogs rename column found_publication_id to publication_id;
alter table dialogs rename column finder_user_id to author_user_id;
alter table dialogs rename column claimer_user_id to respondent_user_id;
