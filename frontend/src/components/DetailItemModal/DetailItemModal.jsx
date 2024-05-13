
import FeatherIcon from 'feather-icons-react'

const DetailItemModal = ({ showDetailItemModal, detailItem, onModalClose, onAddToList }) => {

    return (
        <div className={`${showDetailItemModal ? 'block' : 'hidden'} fixed top-0 left-0 w-screen h-screen z-10 bg-black/80`}
            onClick={onModalClose}
        >
            <div className='h-14 px-6 flex justify-between items-center bg-black/80'
                onClick={(e) => e.stopPropagation()}
            >
                <div className='flex items-baseline gap-x-6'>
                    <h1 className='text-xl text-white'>{detailItem?.name}</h1>
                    <p className='text-white'>Imported date 07-07-2024 17:20</p>
                </div>

                <div className='flex gap-x-3'>
                    <button className='px-3 py-2 sm:py-1 bg-white/20 text-base text-white rounded-full flex items-center gap-x-1 hover:bg-white/30'>
                        <FeatherIcon icon="download" className='w-[18px] h-[18px]' strokeWidth={1.6} />
                        <p className='hidden sm:block'>Download</p>
                    </button>

                    {
                        onAddToList && (
                            <button className='px-3 py-2 sm:py-1 bg-white/20 text-base text-white rounded-full flex items-center gap-x-1 hover:bg-white/30'
                                onClick={onAddToList}
                            >
                                <FeatherIcon icon="plus" className='w-[18px] h-[18px]' strokeWidth={1.6} />
                                <p className='hidden sm:block'>Add to List</p>
                            </button>
                        )
                    }
                </div>
            </div>
        </div>
    )
}

export default DetailItemModal