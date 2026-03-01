import { DocumentDetail } from '@/components/documents/document-detail';

type DocumentPageProps = {
  params: {
    docId: string;
  };
};

export default function DocumentPage({ params }: DocumentPageProps): JSX.Element {
  return <DocumentDetail docId={params.docId} />;
}
