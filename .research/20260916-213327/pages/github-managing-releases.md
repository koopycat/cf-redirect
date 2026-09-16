In this article

# Managing releases in a repository


You can create releases to bundle and deliver iterations of a project to users.

Copy markdown

## Who can use this feature?


Repository collaborators and people with write access to a repository can create, edit, and delete a release.


## Tool navigation


## In this article


## [About release management](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#about-release-management)


You can create new releases with release notes, @mentions of contributors, and links to binary files, as well as edit or delete existing releases. You can also create, modify, and delete releases by using the Releases API. For more information, see [REST API endpoints for releases](https://docs.github.com/en/rest/releases/releases) in the REST API documentation.


You can also publish an action from a specific release in GitHub Marketplace. For more information, see [Publishing actions in GitHub Marketplace](https://docs.github.com/en/actions/how-tos/create-and-publish-actions/publish-in-github-marketplace).


You can choose whether Git Large File Storage (Git LFS) objects are included in the ZIP files and tarballs that GitHub creates for each release. For more information, see [Managing Git LFS objects in archives of your repository](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/managing-git-lfs-objects-in-archives-of-your-repository).


## [Creating a release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#creating-a-release)


Tip


If you have enabled immutable releases for your repository, it's recommended to create releases as drafts first, attach all assets, and then publish. This ensures all assets are in place before the release becomes immutable. For more information, see [Immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases).


1. On GitHub, navigate to the main page of the repository.
2. To the right of the list of files, click **Releases**. ![Screenshot of the main page of a repository. A link, labeled "Releases", is highlighted with an orange outline.](https://docs.github.com/assets/cb-90524/images/help/releases/release-link.png)
3. At the top of the page, click **Draft a new release**.
4. To choose a tag for the release, select the **Choose a tag** dropdown menu.
  - To use an existing tag, click the tag.
  - To create a new tag, type a version number for your release, then click **Create new tag**.5. If you created a new tag, select the **Target** dropdown menu, then click the branch that contains the project you want to release.
6. Optionally, above the description field, select the **Previous tag** dropdown menu, then click the tag that identifies the previous release. ![Screenshot of the "New release" form. A dropdown menu, labeled "Previous tag: auto", is highlighted with an orange outline.](https://docs.github.com/assets/cb-37212/images/help/releases/releases-tag-previous-release.png)
7. In the "Release title" field, type a title for your release.
8. In the "Describe this release" field, type a description for your release. If you @mention anyone in the description, the published release will include a **Contributors** section with an avatar list of all the mentioned users. Alternatively, you can automatically generate your release notes by clicking **Generate release notes**.
9. Optionally, to include binary files such as compiled programs in your release, drag and drop or manually select files in the binaries box.
10. Optionally, to notify users that the release is not ready for production and may be unstable, select **This is a pre-release**.
11. Optionally, select **Set as latest release**. If you do not select this option, the latest release label will automatically be assigned based on semantic versioning.
12. Optionally, if GitHub Discussions is enabled for the repository, create a discussion for the release.
  - Select **Create a discussion for this release**.
  - Select the **Category** dropdown menu, then click a category for the release discussion.13. If you're ready to publicize your release, click **Publish release**. To work on the release later, click **Save draft**. If you have enabled immutable releases for the repository, creating a draft first allows you to attach all assets before the release becomes immutable. You can then view your published or draft releases in the releases feed for your repository. For more information, see [Viewing your repository's releases and tags](https://docs.github.com/en/repositories/releasing-projects-on-github/viewing-your-repositorys-releases-and-tags).


Note


To learn more about GitHub CLI, see [About GitHub CLI](https://docs.github.com/en/github-cli/github-cli/about-github-cli).


1. To create a release, use the `gh release create` subcommand. Replace `tag` with the desired tag for the release. ```bash gh release create TAG ```
2. Follow the interactive prompts. Alternatively, you can specify arguments to skip these prompts. For more information about possible arguments, see [the GitHub CLI manual](https://cli.github.com/manual/gh_release_create). For example, this command creates a prerelease with the specified title and notes.


```bash
gh release create v1.3.2 --title "v1.3.2 (beta)" --notes "this is a public preview release" --prerelease
```


If you @mention any GitHub users in the notes, the published release will include a **Contributors** section with an avatar list of all the mentioned users.


## [Editing a release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#editing-a-release)


Note


If you have enabled immutable releases for your repository, you cannot add, replace, or delete assets after a release is published, and you cannot move or delete its tag while the release exists. You can still edit the title and release notes, and change whether the release is a pre-release or the latest release. See [Immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases).


1. On GitHub, navigate to the main page of the repository.
2. To the right of the list of files, click **Releases**. ![Screenshot of the main page of a repository. A link, labeled "Releases", is highlighted with an orange outline.](https://docs.github.com/assets/cb-90524/images/help/releases/release-link.png)
3. Next to the release you want to edit, click . ![Screenshot of a release in the releases list. A pencil icon is highlighted with an orange outline.](https://docs.github.com/assets/cb-33857/images/help/releases/edit-release-pencil.png)
4. Edit the details for the release in the form, then click **Update release**. If you add or remove any @mentions of GitHub users in the description, those users will be added or removed from the avatar list in the **Contributors** section of the release.


1. To edit a release, use the `gh release edit` subcommand. Replace `TAG` with the tag representing the release you wish to edit. For example, to edit the title for a release, use the following code, replacing `NEW-TITLE` with the updated title: ```bash gh release edit TAG -t "NEW-TITLE" ``` For more information about possible arguments, see [the GitHub CLI manual](https://cli.github.com/manual/gh_release_edit).


## [Deleting a release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#deleting-a-release)


1. On GitHub, navigate to the main page of the repository.
2. To the right of the list of files, click **Releases**. ![Screenshot of the main page of a repository. A link, labeled "Releases", is highlighted with an orange outline.](https://docs.github.com/assets/cb-90524/images/help/releases/release-link.png)
3. On the right side of the page, next to the release you want to delete, click . ![Screenshot of a release in the releases list. A trash icon is highlighted with an orange outline.](https://docs.github.com/assets/cb-33811/images/help/releases/delete-release-trash.png)
4. Click **Delete this release**.


1. To delete a release, use the `gh release delete` subcommand. Replace `tag` with the tag of the release to delete. Use the `-y` flag to skip confirmation. ```bash gh release delete TAG -y ```[Make a contribution](https://github.com/github/docs/blob/main/content/repositories/releasing-projects-on-github/managing-releases-in-a-repository.md)

[Terms](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service)

[Privacy](https://docs.github.com/en/site-policy/privacy-policies/github-privacy-statement)
