import { Repository } from '@browser-git/browser-git';

let repo: Repository | null = null;

// Helper functions for UI
function showStatus(message: string, type: 'success' | 'error' | 'info') {
  const statusDiv = document.getElementById('status')!;
  statusDiv.className = `status ${type}`;
  statusDiv.textContent = message;
  setTimeout(() => {
    statusDiv.className = '';
    statusDiv.textContent = '';
  }, 5000);
}

function appendOutput(text: string) {
  const outputDiv = document.getElementById('output')!;
  outputDiv.textContent += text + '\n';
  outputDiv.scrollTop = outputDiv.scrollHeight;
}

function clearOutput() {
  const outputDiv = document.getElementById('output')!;
  outputDiv.textContent = '';
}

// Repository operations
(window as any).initRepo = async () => {
  try {
    clearOutput();
    appendOutput('Initializing repository...');

    repo = await Repository.init('/demo-repo', {
      storage: 'indexeddb',
      hashAlgorithm: 'sha1',
      initialBranch: 'main',
    });

    appendOutput('✓ Repository initialized successfully!');
    appendOutput(`Storage: IndexedDB`);
    appendOutput(`Hash Algorithm: SHA-1`);
    appendOutput(`Initial Branch: main`);

    showStatus('Repository initialized successfully!', 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).getStatus = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    clearOutput();
    appendOutput('Getting repository status...\n');

    const status = await repo.status();

    appendOutput(`Branch: ${status.branch || 'HEAD'}\n`);

    if (status.added.length > 0) {
      appendOutput('Staged files:');
      status.added.forEach(file => appendOutput(`  ✓ ${file}`));
      appendOutput('');
    }

    if (status.modified.length > 0) {
      appendOutput('Modified files:');
      status.modified.forEach(file => appendOutput(`  M ${file}`));
      appendOutput('');
    }

    if (status.deleted.length > 0) {
      appendOutput('Deleted files:');
      status.deleted.forEach(file => appendOutput(`  D ${file}`));
      appendOutput('');
    }

    if (status.untracked.length > 0) {
      appendOutput('Untracked files:');
      status.untracked.forEach(file => appendOutput(`  ? ${file}`));
      appendOutput('');
    }

    if (status.added.length === 0 && status.modified.length === 0 &&
        status.deleted.length === 0 && status.untracked.length === 0) {
      appendOutput('✓ Working tree clean');
    }

    showStatus('Status retrieved successfully!', 'info');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

// File operations
(window as any).createFile = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    const filename = (document.getElementById('filename') as HTMLInputElement).value;
    const content = (document.getElementById('filecontent') as HTMLTextAreaElement).value;

    clearOutput();
    appendOutput(`Creating file: ${filename}...`);

    await repo.fs.writeFile(filename, content);

    appendOutput(`✓ File created: ${filename}`);
    appendOutput(`Size: ${content.length} bytes`);

    showStatus(`File ${filename} created successfully!`, 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).readFile = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    const filename = (document.getElementById('filename') as HTMLInputElement).value;

    clearOutput();
    appendOutput(`Reading file: ${filename}...\n`);

    const content = await repo.fs.readFile(filename, 'utf-8');

    appendOutput('File contents:');
    appendOutput('─'.repeat(50));
    appendOutput(content);
    appendOutput('─'.repeat(50));

    showStatus(`File ${filename} read successfully!`, 'info');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

// Git operations
(window as any).addFiles = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    clearOutput();
    appendOutput('Adding files to staging area...');

    await repo.add(['.']);

    appendOutput('✓ All files added to staging area');

    showStatus('Files added successfully!', 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).commitChanges = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    const message = (document.getElementById('commitMsg') as HTMLInputElement).value;

    clearOutput();
    appendOutput('Creating commit...');

    const hash = await repo.commit(message, {
      author: {
        name: 'Demo User',
        email: 'demo@browsergit.dev',
      },
    });

    appendOutput(`✓ Commit created: ${hash.substring(0, 7)}`);
    appendOutput(`Message: ${message}`);

    showStatus('Commit created successfully!', 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).viewLog = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    clearOutput();
    appendOutput('Commit History:\n');

    const commits = await repo.log({ maxCount: 20 });

    if (commits.length === 0) {
      appendOutput('No commits yet');
      return;
    }

    commits.forEach((commit, index) => {
      appendOutput(`commit ${commit.hash}`);
      appendOutput(`Author: ${commit.author.name} <${commit.author.email}>`);
      appendOutput(`Date:   ${new Date(commit.date).toLocaleString()}`);
      appendOutput('');
      commit.message.split('\n').forEach(line => {
        appendOutput(`    ${line}`);
      });
      if (index < commits.length - 1) {
        appendOutput('\n' + '─'.repeat(50) + '\n');
      }
    });

    showStatus('Log retrieved successfully!', 'info');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

// Branch operations
(window as any).createBranch = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    const branchName = (document.getElementById('branchName') as HTMLInputElement).value;

    clearOutput();
    appendOutput(`Creating branch: ${branchName}...`);

    await repo.createBranch(branchName);

    appendOutput(`✓ Branch created: ${branchName}`);

    showStatus(`Branch ${branchName} created successfully!`, 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).listBranches = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    clearOutput();
    appendOutput('Branches:\n');

    const branches = await repo.listBranches();
    const currentBranch = await repo.currentBranch();

    branches.forEach(branch => {
      if (branch === currentBranch) {
        appendOutput(`* ${branch} (current)`);
      } else {
        appendOutput(`  ${branch}`);
      }
    });

    showStatus('Branches listed successfully!', 'info');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

(window as any).checkoutBranch = async () => {
  try {
    if (!repo) {
      repo = await Repository.open('/demo-repo');
    }

    const branchName = (document.getElementById('branchName') as HTMLInputElement).value;

    clearOutput();
    appendOutput(`Checking out branch: ${branchName}...`);

    await repo.checkout(branchName);

    appendOutput(`✓ Switched to branch: ${branchName}`);

    showStatus(`Switched to branch ${branchName}!`, 'success');
  } catch (error) {
    appendOutput(`✗ Error: ${(error as Error).message}`);
    showStatus(`Error: ${(error as Error).message}`, 'error');
  }
};

// Initialize on load
appendOutput('Welcome to BrowserGit Demo!');
appendOutput('Click "Initialize Repository" to get started.\n');
