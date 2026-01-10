import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  icon: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Full Git in the Browser',
    icon: '🌐',
    description: (
      <>
        Complete Git implementation running entirely client-side. Clone, commit,
        push, pull, merge, and more - all without a server for local operations.
      </>
    ),
  },
  {
    title: 'High Performance',
    icon: '⚡',
    description: (
      <>
        Built with Go + WebAssembly for native-like performance. Commits in under
        1ms, checkouts in under 1ms, and optimized for large repositories.
      </>
    ),
  },
  {
    title: 'Flexible Storage',
    icon: '💾',
    description: (
      <>
        Multiple storage backends: IndexedDB, OPFS, LocalStorage, or in-memory.
        Automatic feature detection with graceful fallbacks.
      </>
    ),
  },
];

function Feature({title, icon, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <span className={styles.featureIcon} role="img">{icon}</span>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
