# QA Procedure Template

For the human-reviewed tier of the gauntlet: the human reads this
document and the linked gherkin/acceptance criteria, not the
implementation or unit tests. Fill in one copy per scenario.

## Scenario name

<!-- Short, specific name. Match the gherkin Scenario title where one exists. -->

## Preconditions

<!-- State the system must be in before starting: data, config, feature
flags, environment. List each as a separate line. -->

-
-

## Manual steps

1.
2.
3.

## Expected result

<!-- What the reviewer should observe after the last step. Be specific
enough that "it worked" and "it didn't" are unambiguous. -->

## Rollback / abort steps

<!-- How to return the system to its precondition state if the result is
wrong, or if the procedure must be stopped partway through. -->

1.
2.

## Sign-off

Reviewed by: __________________________  Date: ______________

Result: [ ] Pass    [ ] Fail    [ ] Blocked
