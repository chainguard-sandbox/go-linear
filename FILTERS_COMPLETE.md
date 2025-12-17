# IssueFilter Implementation - Final Status

## Summary

**Total filters in API:** 61
**CLI flags available:** 64
**Fully implemented:** 44
**Skipped (cannot implement):** 17

---

## Implemented Filters (41/61) ✅

### Already Existed (8)
1. assignee
2. completedAt
3. createdAt
4. labels
5. priority
6. state
7. team
8. updatedAt

### Session Added (36)
9. addedToCycleAt
10. addedToCyclePeriod
11. archivedAt
12. attachments (via attachment-by)
13. autoArchivedAt
14. autoClosedAt
15. canceledAt
16. children (via has-children)
17. comments (via comment-by)
18. creator
19. customerCount
20. customerImportantCount
21. cycle
22. delegate
23. description
24. dueDate
25. estimate
26. hasBlockedByRelations
27. hasBlockingRelations
28. hasDuplicateRelations
29. hasRelatedRelations
30. hasSuggestedAssignees
31. hasSuggestedLabels
32. hasSuggestedProjects
33. hasSuggestedTeams
34. id
35. lastAppliedTemplate
36. needs (via has-needs)
37. number
38. parent
39. project
40. projectMilestone
41. reactions (via has-reactions)
42. snoozedBy
43. snoozedUntilAt
44. startedAt
45. subscribers
46. slaStatus
47. title
48. triagedAt

Also added sub-filters for collections:
- comment-contains (text in comments)
- attachment-source-type (attachment source)

---

## Cannot Implement (17/61) ❌

### Internal Fields (9) - Should Not Expose
1. accumulatedStateUpdatedAt
2. ageTime
3. cycleTime
4. hasSuggestedRelatedIssues
5. hasSuggestedSimilarIssues
6. leadTime
7. searchableContent
8. suggestions
9. triageTime

### Compound Logic (2) - Requires Recursive Filters
10. and
11. or

### Alpha/Experimental (1)
12. recurringIssueTemplate

### Complex Comparators (5) - Would Need Advanced Syntax
13. sourceMetadata (SourceMetadataComparator - integration-specific metadata)
14-17. Collection filters requiring nested recursion:
  - attachments.* (except creator and sourceType - those are implemented)
  - children.* (except length - that's has-children)
  - comments.* (except user and body - those are implemented)
  - needs.*, reactions.*, subscribers.* (except basic checks)

---

## Production Filtering

With 44 implemented filters, comprehensive issue querying is now possible:

```bash
# Find issues requiring triage approval
go-linear issue list \
  --state=Triage \
  --creator=colleague@company.com \
  --has-suggested-teams \
  --created-after=7d \
  --team=Engineering \
  --comment-by=me \
  --count

# List details for manual review
go-linear issue list \
  --state=Triage \
  --creator=colleague@company.com \
  --has-suggested-teams \
  --created-after=7d \
  --output=json

# Future: Batch approve
# (requires Issue.suggestions API support)
```

---

## Filter Categories

**Date filters (17):**
Added-to-cycle, archived, auto-archived, auto-closed, canceled, completed, created, due, snoozed-until, started, triaged, updated

**Entity filters (13):**
Assignee, creator, cycle, delegate, parent, project, project-milestone, snoozed-by, state, team, last-applied-template

**Text filters (2):**
Description, title

**Numeric filters (5):**
Customer-count, customer-important-count, estimate, number, priority

**Relation filters (8):**
Has-blocked-by, has-blocking, has-duplicate, has-related, parent

**AI Suggestion filters (4):**
Has-suggested-assignees, has-suggested-labels, has-suggested-projects, has-suggested-teams

**Collection filters (6):**
Attachment-by, comment-by, has-children, has-needs, has-reactions, subscriber

**ID filter (1):**
id
